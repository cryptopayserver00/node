package core

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/sweep/setup"
	"node/utils"
	"node/utils/notification"
	"slices"
	"strconv"
	"sync"
	"time"

	sweepUtils "node/sweep/utils"

	"github.com/gagliardetto/solana-go"
	lookup "github.com/gagliardetto/solana-go/programs/address-lookup-table"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/redis/go-redis/v9"
)

func SetupSolLatestBlockHeight(ctx context.Context, chainId uint) {
	endpoint := constant.GetRPCUrlByNetwork(chainId)
	client := rpc.New(endpoint)

	height, err := client.GetSlot(ctx, rpc.CommitmentFinalized)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return
	}

	if height > 0 {
		setup.SetupLatestBlockHeight(ctx, chainId, int64(height))
	}
}

func SweepSolBlockchainTransaction(
	ctx context.Context,
	chainId uint,
	publicKey *[]string,
	sweepBlockHeight, cacheBlockHeight *int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) {
	defer utils.HandlePanic()

	if len(*publicKey) <= 0 {
		SetupSolLatestBlockHeight(ctx, chainId)
		setup.UpdateCacheBlockHeight(ctx, chainId)
		setup.UpdateSweepBlockHeight(ctx, chainId)
		setup.UpdatePublicKey(ctx, chainId)
		return
	}

	if *sweepBlockHeight >= *cacheBlockHeight {
		SetupSolLatestBlockHeight(ctx, chainId)
		setup.UpdateCacheBlockHeight(ctx, chainId)
		setup.UpdatePublicKey(ctx, chainId)
		time.Sleep(time.Second * 1)
		return
	}

	var wg sync.WaitGroup

	numWorkers := calcNumWorkers(*sweepBlockHeight, *cacheBlockHeight, maxWorkers)
	if numWorkers == 0 {
		return
	}

	start := *sweepBlockHeight
	end := start + int64(numWorkers) - 1

	for h := start; h <= end; h++ {
		wg.Add(1)
		go func(height int64) {
			defer wg.Done()
			defer utils.HandlePanic()

			err := SweepSolBlockchainTransactionCore(ctx, chainId, publicKey, height, constantSweepBlock, constantPendingBlock, constantPendingTransaction)
			if err != nil {
				if _, rpushErr := global.NODE_REDIS.RPush(ctx, constantPendingBlock, height).Result(); rpushErr != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), rpushErr.Error()))
				}
			}
		}(h)
	}

	wg.Wait()
	*sweepBlockHeight = end + 1
	if _, err := global.NODE_REDIS.Set(ctx, constantSweepBlock, *sweepBlockHeight, 0).Result(); err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}
}

func SweepSolBlockchainTransactionCore(
	ctx context.Context,
	chainId uint,
	publicKey *[]string,
	sweepBlockHeight int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) error {
	defer utils.HandlePanic()

	var err error

	endpoint := constant.GetRPCUrlByNetwork(chainId)
	client := rpc.New(endpoint)

	includeRewards := false

	blockResult, err := client.GetBlockWithOpts(ctx, uint64(sweepBlockHeight), &rpc.GetBlockOpts{
		Encoding:                       solana.EncodingBase64,
		Commitment:                     rpc.CommitmentConfirmed,
		Rewards:                        &includeRewards,
		MaxSupportedTransactionVersion: &rpc.MaxSupportedTransactionVersion1,
	})

	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	if blockResult.BlockHeight == nil || blockResult.ParentSlot == 0 {
		err = errors.New("block height is nil in RPC response")
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	if uint64(sweepBlockHeight) == blockResult.ParentSlot+1 {
		if len(blockResult.Transactions) > 0 {
			for _, transaction := range blockResult.Transactions {

				if transaction.Meta != nil && transaction.Meta.Err != nil {
					continue
				}

				isMonitorTx := false
				matchArray := make([]solana.PublicKey, 0)

				parsedTx, err := transaction.GetTransaction()
				if err != nil {
					continue
				}

				if len(parsedTx.Signatures) == 0 {
					continue
				}

				if err := processTransactionWithAddressLookups(ctx, client, parsedTx); err != nil {
					continue
				}

				allAccountKeys := parsedTx.Message.AccountKeys

				for _, inst := range parsedTx.Message.Instructions {

					if int(inst.ProgramIDIndex) >= len(allAccountKeys) {
						continue
					}
					programID := allAccountKeys[inst.ProgramIDIndex]

					switch programID {
					case solana.SystemProgramID:
						if len(inst.Data) >= 12 && binary.LittleEndian.Uint32(inst.Data[0:4]) == 2 {
							if len(inst.Accounts) < 2 {
								continue
							}
							from := allAccountKeys[inst.Accounts[0]]
							to := allAccountKeys[inst.Accounts[1]]

							matchArray = append(matchArray, from, to)
						}

					case solana.TokenProgramID:
						if len(inst.Data) < 1 {
							continue
						}

						switch inst.Data[0] {
						case 3:
							// Transfer
							if len(inst.Data) < 9 || len(inst.Accounts) < 2 {
								continue
							}

							fromIdx := inst.Accounts[0]
							toIdx := inst.Accounts[1]

							fromOwner, fromMint, fromOk := lookupTokenAccountOwnerMint(transaction.Meta, fromIdx)
							toOwner, toMint, toOk := lookupTokenAccountOwnerMint(transaction.Meta, toIdx)

							if !fromOk || !toOk {
								continue // 查不到说明这不是一次标准 SPL token 转账（或 meta 数据缺失），跳过
							}

							if !fromMint.Equals(toMint) {
								continue
							}

							isSupportContract, _, _, _ := sweepUtils.GetContractInfo(chainId, fromMint.String())
							if !isSupportContract {
								continue
							}

							matchArray = append(matchArray, fromOwner, toOwner)
						case 12:
							// TransferChecked
							if len(inst.Accounts) < 3 {
								continue
							}

							fromIdx := inst.Accounts[0]
							// inst.Accounts[1] 是 mint account，TransferChecked 指令本身带了 mint，不需要再查
							toIdx := inst.Accounts[2]

							fromOwner, fromMint, fromOk := lookupTokenAccountOwnerMint(transaction.Meta, fromIdx)
							toOwner, _, toOk := lookupTokenAccountOwnerMint(transaction.Meta, toIdx)

							if !fromOk || !toOk {
								continue
							}

							isSupportContract, _, _, _ := sweepUtils.GetContractInfo(chainId, fromMint.String())
							if !isSupportContract {
								continue
							}

							matchArray = append(matchArray, fromOwner, toOwner)
						default:
							continue
						}
					}
				}

				if len(matchArray) == 0 {
					continue
				}

				matchArray = utils.RemoveDuplicatesForSolanaPublicKey(matchArray)

			outerCurrentTxLoop:
				for i := 0; i < len(*publicKey); i++ {
					targetAddress := solana.MustPublicKeyFromBase58((*publicKey)[i])
					for _, j := range matchArray {
						if targetAddress.Equals(j) {
							isMonitorTx = true
							break outerCurrentTxLoop
						}
					}
				}

				if isMonitorTx {
					redisTxs, err := global.NODE_REDIS.LRange(ctx, constantPendingTransaction, 0, -1).Result()
					if err != nil {
						global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
						time.Sleep(2 * time.Second)
						return err
					}

					sign := parsedTx.Signatures[0].String()

					found := slices.Contains(redisTxs, sign)

					if !found {
						_, err = global.NODE_REDIS.RPush(ctx, constantPendingTransaction, sign).Result()
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							time.Sleep(2 * time.Second)
							return err
						}
					}
				}
			}
		}

		return nil
	} else {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", sweepBlockHeight, blockResult.ParentSlot+1)))
		time.Sleep(2 * time.Second)
		return errors.New("not the same height of block")
	}
}

func SweepSolBlockchainTransactionDetails(
	ctx context.Context,
	chainId uint,
	publicKey *[]string,
	txRetryCount *map[string]int,
	constantPendingTransaction string,
) {
	defer utils.HandlePanic()

	txHash, err := global.NODE_REDIS.LIndex(ctx, constantPendingTransaction, 0).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			time.Sleep(2 * time.Second)
			return
		}
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return
	}

	global.NODE_LOG.Info(fmt.Sprintf("%s -> handle tx: %s", constant.GetChainName(chainId), txHash))

	err = handleSolTransactionDetails(ctx, chainId, publicKey, txHash)

	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("Can not handle the tx: %s, Retry | %s -> %s", txHash, constant.GetChainName(chainId), err.Error()))

		retryCount := (*txRetryCount)[txHash]
		retryCount++
		if retryCount >= setup.SweepThreshold {
			global.NODE_LOG.Error(fmt.Sprintf("give up tx after retries: %s", txHash))
			_, err = global.NODE_REDIS.LPop(ctx, constantPendingTransaction).Result()
			if err != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				time.Sleep(2 * time.Second)
				return
			}
			delete(*txRetryCount, txHash)
			return
		}
		(*txRetryCount)[txHash] = retryCount
		time.Sleep(2 * time.Second)
		return
	}

	_, err = global.NODE_REDIS.LPop(ctx, constantPendingTransaction).Result()
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return
	}
	delete(*txRetryCount, txHash)
}

func handleSolTransactionDetails(
	ctx context.Context,
	chainId uint,
	publicKey *[]string,
	txHash string,
) error {
	endpoint := constant.GetRPCUrlByNetwork(chainId)
	client := rpc.New(endpoint)

	txSig := solana.MustSignatureFromBase58(txHash)

	transactionResult, err := client.GetTransaction(ctx, txSig, &rpc.GetTransactionOpts{
		MaxSupportedTransactionVersion: &rpc.MaxSupportedTransactionVersion1,
		Encoding:                       solana.EncodingBase64,
	})

	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	if transactionResult.Meta != nil && transactionResult.Meta.Err != nil {
		return errors.New("transaction failed on-chain")
	}

	transaction, err := transactionResult.Transaction.GetTransaction()
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	if len(transaction.Signatures) == 0 {
		err = errors.New("transaction signatures length can not be 0")
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return err
	}

	if err := processTransactionWithAddressLookups(ctx, client, transaction); err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> address lookup resolve failed: %s", constant.GetChainName(chainId), err.Error()))
		return err
	}

	var notifyRequest request.NotificationRequest

	notifyRequest.Hash = transaction.Signatures[0].String()
	notifyRequest.Chain = chainId

	if transactionResult.BlockTime != nil {
		notifyRequest.BlockTimestamp = int(transactionResult.BlockTime.Time().Unix()) * 1000
	}

	isProcess := false

	for _, inst := range transaction.Message.Instructions {

		if int(inst.ProgramIDIndex) >= len(transaction.Message.AccountKeys) {
			continue
		}
		programID := transaction.Message.AccountKeys[inst.ProgramIDIndex]

		switch programID {
		case solana.SystemProgramID:
			if len(inst.Data) >= 12 && binary.LittleEndian.Uint32(inst.Data[0:4]) == 2 {
				if len(inst.Accounts) < 2 {
					continue
				}

				amount := binary.LittleEndian.Uint64(inst.Data[4:12])
				from := transaction.Message.AccountKeys[inst.Accounts[0]]
				to := transaction.Message.AccountKeys[inst.Accounts[1]]

				notifyRequest.Token = "SOL"
				notifyRequest.FromAddress = from.String()
				notifyRequest.ToAddress = to.String()
				notifyRequest.Amount = utils.CalculateBalance(big.NewInt(int64(amount)), 9)

				for _, v := range *publicKey {
					targetAddress := solana.MustPublicKeyFromBase58(v)

					if targetAddress.Equals(from) {
						notifyRequest.TransactType = "send"
						notifyRequest.Address = v

						err = notification.NotificationRequest(ctx, notifyRequest)
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							return err
						}
						isProcess = true
					}

					if targetAddress.Equals(to) {
						notifyRequest.TransactType = "receive"
						notifyRequest.Address = v

						err = notification.NotificationRequest(ctx, notifyRequest)
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							return err
						}
						isProcess = true
					}
				}

			}

		case solana.TokenProgramID:
			if len(inst.Data) < 1 {
				continue
			}

			var from, to solana.PublicKey

			switch inst.Data[0] {
			case 3:
				if len(inst.Data) < 9 || len(inst.Accounts) < 2 {
					continue
				}

				amount := binary.LittleEndian.Uint64(inst.Data[1:9])
				fromIdx := inst.Accounts[0]
				toIdx := inst.Accounts[1]

				fromOwner, fromMint, fromOk := lookupTokenAccountOwnerMint(transactionResult.Meta, fromIdx)
				toOwner, toMint, toOk := lookupTokenAccountOwnerMint(transactionResult.Meta, toIdx)

				if !fromOk || !toOk {
					continue
				}

				if !fromMint.Equals(toMint) {
					continue
				}

				isSupportContract, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, fromMint.String())
				if !isSupportContract {
					continue
				}

				notifyRequest.FromAddress = fromOwner.String()
				notifyRequest.ToAddress = toOwner.String()
				notifyRequest.Token = contractName
				notifyRequest.Amount = utils.CalculateBalance(big.NewInt(int64(amount)), decimals)
				from = fromOwner
				to = toOwner
			case 12:
				if len(inst.Data) < 9 || len(inst.Accounts) < 3 {
					continue
				}

				amount := binary.LittleEndian.Uint64(inst.Data[1:9])
				fromIdx := inst.Accounts[0]
				toIdx := inst.Accounts[2]

				fromOwner, fromMint, fromOk := lookupTokenAccountOwnerMint(transactionResult.Meta, fromIdx)
				toOwner, _, toOk := lookupTokenAccountOwnerMint(transactionResult.Meta, toIdx)

				if !fromOk || !toOk {
					continue
				}

				isSupportContract, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, fromMint.String())
				if !isSupportContract {
					continue
				}

				notifyRequest.FromAddress = fromOwner.String()
				notifyRequest.ToAddress = toOwner.String()
				notifyRequest.Token = contractName
				notifyRequest.Amount = utils.CalculateBalance(big.NewInt(int64(amount)), decimals)
				from = fromOwner
				to = toOwner
			default:
				continue
			}

			for _, v := range *publicKey {
				targetAddress := solana.MustPublicKeyFromBase58(v)

				if targetAddress.Equals(from) {
					notifyRequest.TransactType = "send"
					notifyRequest.Address = v

					err = notification.NotificationRequest(ctx, notifyRequest)
					if err != nil {
						global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
						return err
					}
					isProcess = true
				}

				if targetAddress.Equals(to) {
					notifyRequest.TransactType = "receive"
					notifyRequest.Address = v

					err = notification.NotificationRequest(ctx, notifyRequest)
					if err != nil {
						global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
						return err
					}
					isProcess = true
				}
			}
		default:
			continue
		}
	}

	if !isProcess {
		return fmt.Errorf("no monitored address matched for tx: %s", notifyRequest.Hash)
	}

	return nil
}

func SweepSolBlockchainPendingBlock(
	ctx context.Context,
	chainId uint,
	publicKey *[]string,
	retryBlockCount *map[int64]int,
	constantPendingBlock, constantPendingTransaction string,
) {
	defer utils.HandlePanic()

	blockHeight, err := global.NODE_REDIS.LIndex(ctx, constantPendingBlock, 0).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			time.Sleep(2 * time.Second)
			return
		}
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(10 * time.Second)
		return
	}

	blockHeightInt, err := strconv.ParseInt(blockHeight, 10, 64)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	endpoint := constant.GetRPCUrlByNetwork(chainId)
	client := rpc.New(endpoint)

	includeRewards := false

	blockResult, err := client.GetBlockWithOpts(ctx, uint64(blockHeightInt), &rpc.GetBlockOpts{
		Encoding:                       solana.EncodingBase64,
		Commitment:                     rpc.CommitmentConfirmed,
		Rewards:                        &includeRewards,
		MaxSupportedTransactionVersion: &rpc.MaxSupportedTransactionVersion1,
	})

	if err != nil {
		time.Sleep(2 * time.Second)
		return
	}

	if blockResult.BlockHeight == nil || blockResult.ParentSlot == 0 {
		err = errors.New("block height is nil in RPC response")
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return
	}

	if uint64(blockHeightInt) == blockResult.ParentSlot+1 {
		if len(blockResult.Transactions) > 0 {
			for _, transaction := range blockResult.Transactions {

				if transaction.Meta != nil && transaction.Meta.Err != nil {
					continue
				}

				isMonitorTx := false
				matchArray := make([]solana.PublicKey, 0)

				parsedTx, err := transaction.GetTransaction()
				if err != nil {
					continue
				}

				if len(parsedTx.Signatures) == 0 {
					continue
				}

				if err := processTransactionWithAddressLookups(ctx, client, parsedTx); err != nil {
					continue
				}

				allAccountKeys := parsedTx.Message.AccountKeys

				for _, inst := range parsedTx.Message.Instructions {

					if int(inst.ProgramIDIndex) >= len(allAccountKeys) {
						continue
					}
					programID := allAccountKeys[inst.ProgramIDIndex]

					switch programID {
					case solana.SystemProgramID:
						if len(inst.Data) >= 12 && binary.LittleEndian.Uint32(inst.Data[0:4]) == 2 {
							if len(inst.Accounts) < 2 {
								continue
							}

							from := allAccountKeys[inst.Accounts[0]]
							to := allAccountKeys[inst.Accounts[1]]

							matchArray = append(matchArray, from, to)
						}

					case solana.TokenProgramID:
						if len(inst.Data) < 1 {
							continue
						}

						switch inst.Data[0] {
						case 3:
							if len(inst.Data) < 9 || len(inst.Accounts) < 2 {
								continue
							}

							fromIdx := inst.Accounts[0]
							toIdx := inst.Accounts[1]

							fromOwner, fromMint, fromOk := lookupTokenAccountOwnerMint(transaction.Meta, fromIdx)
							toOwner, toMint, toOk := lookupTokenAccountOwnerMint(transaction.Meta, toIdx)

							if !fromOk || !toOk {
								continue
							}

							if !fromMint.Equals(toMint) {
								continue
							}

							isSupportContract, _, _, _ := sweepUtils.GetContractInfo(chainId, fromMint.String())
							if !isSupportContract {
								continue
							}

							matchArray = append(matchArray, fromOwner, toOwner)
						case 12:
							if len(inst.Accounts) < 3 {
								continue
							}

							fromIdx := inst.Accounts[0]
							toIdx := inst.Accounts[2]

							fromOwner, fromMint, fromOk := lookupTokenAccountOwnerMint(transaction.Meta, fromIdx)
							toOwner, _, toOk := lookupTokenAccountOwnerMint(transaction.Meta, toIdx)

							if !fromOk || !toOk {
								continue
							}

							isSupportContract, _, _, _ := sweepUtils.GetContractInfo(chainId, fromMint.String())
							if !isSupportContract {
								continue
							}

							matchArray = append(matchArray, fromOwner, toOwner)
						default:
							continue
						}
					}
				}

				if len(matchArray) == 0 {
					continue
				}

				matchArray = utils.RemoveDuplicatesForSolanaPublicKey(matchArray)

			outerCurrentTxLoop:
				for i := 0; i < len(*publicKey); i++ {
					targetAddress := solana.MustPublicKeyFromBase58((*publicKey)[i])
					for _, j := range matchArray {
						if targetAddress.Equals(j) {
							isMonitorTx = true
							break outerCurrentTxLoop
						}
					}
				}

				if isMonitorTx {
					redisTxs, err := global.NODE_REDIS.LRange(ctx, constantPendingTransaction, 0, -1).Result()
					if err != nil {
						global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
						time.Sleep(2 * time.Second)
						return
					}

					sign := parsedTx.Signatures[0].String()

					found := slices.Contains(redisTxs, sign)

					if !found {
						_, err = global.NODE_REDIS.RPush(ctx, constantPendingTransaction, sign).Result()
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							time.Sleep(2 * time.Second)
							return
						}
					}
				}
			}
		}

		_, err = global.NODE_REDIS.LPop(ctx, constantPendingBlock).Result()
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			time.Sleep(2 * time.Second)
			return
		}
		delete(*retryBlockCount, blockHeightInt)
	} else {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", blockHeightInt, blockResult.ParentSlot+1)))

		retryCount := (*retryBlockCount)[blockHeightInt]
		retryCount++
		if retryCount >= setup.SweepThreshold {
			global.NODE_LOG.Error(fmt.Sprintf("give up block after retries: %d", blockHeightInt))
			_, err = global.NODE_REDIS.LPop(ctx, constantPendingBlock).Result()
			if err != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				time.Sleep(2 * time.Second)
				return
			}
			delete(*retryBlockCount, blockHeightInt)
			return
		}
		(*retryBlockCount)[blockHeightInt] = retryCount
		time.Sleep(2 * time.Second)
	}
}

func processTransactionWithAddressLookups(ctx context.Context, rpc *rpc.Client, tx *solana.Transaction) error {
	if !tx.Message.IsVersioned() {
		return nil
	}

	tableKeys := tx.Message.GetAddressTableLookups().GetTableIDs()
	if len(tableKeys) == 0 {
		return nil
	}

	if tx.Message.GetAddressTableLookups().NumLookups() == 0 {
		return nil
	}

	resolutions := make(map[solana.PublicKey]solana.PublicKeySlice)
	for _, key := range tableKeys {
		info, err := rpc.GetAccountInfo(ctx, key)
		if err != nil {
			return fmt.Errorf("get address lookup table account info: %w", err)
		}
		tableContent, err := lookup.DecodeAddressLookupTableState(info.GetBinary())
		if err != nil {
			return fmt.Errorf("decode address lookup table: %w", err)
		}
		resolutions[key] = tableContent.Addresses
	}

	if err := tx.Message.SetAddressTables(resolutions); err != nil {
		return fmt.Errorf("set address tables: %w", err)
	}

	if err := tx.Message.ResolveLookups(); err != nil {
		return fmt.Errorf("resolve lookups: %w", err)
	}

	return nil
}

// lookupTokenAccountOwnerMint 从交易 Meta 的 PostTokenBalances / PreTokenBalances 中
// 按账户在 AccountKeys 里的索引查找 owner 和 mint，不需要额外发 RPC。
// 优先用 Post（转账后的状态更可靠，尤其是接收方账户可能是本次交易里才创建的），
// 查不到再退回 Pre（适合账户在转账后被关闭的情况，比如全额转出后 close 掉）。
func lookupTokenAccountOwnerMint(meta *rpc.TransactionMeta, accountIndex uint16) (owner, mint solana.PublicKey, ok bool) {
	if meta == nil {
		return
	}

	for _, tb := range meta.PostTokenBalances {
		if tb.AccountIndex == accountIndex {
			if tb.Owner == nil {
				return
			}
			return *tb.Owner, tb.Mint, true
		}
	}

	for _, tb := range meta.PreTokenBalances {
		if tb.AccountIndex == accountIndex {
			if tb.Owner == nil {
				return
			}
			return *tb.Owner, tb.Mint, true
		}
	}

	return
}
