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
	"strconv"
	"sync"
	"time"

	sweepUtils "node/sweep/utils"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/redis/go-redis/v9"
)

func SetupSolLatestBlockHeight(chainId uint) {
	endpoint := constant.GetRPCUrlByNetwork(chainId)
	client := rpc.New(endpoint)

	height, err := client.GetSlot(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	if height > 0 {
		setup.SetupLatestBlockHeight(context.Background(), chainId, int64(height))
	}
}

func SweepSolBlockchainTransaction(
	chainId uint,
	publicKey *[]string,
	sweepCount *map[int64]int,
	sweepBlockHeight, cacheBlockHeight *int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) {
	defer utils.HandlePanic()

	if len(*publicKey) <= 0 {
		SetupSolLatestBlockHeight(chainId)
		setup.UpdateCacheBlockHeight(context.Background(), chainId)
		setup.UpdateSweepBlockHeight(context.Background(), chainId)
		setup.UpdatePublicKey(context.Background(), chainId)
		return
	}

	if *sweepBlockHeight >= *cacheBlockHeight {
		SetupSolLatestBlockHeight(chainId)
		setup.UpdateCacheBlockHeight(context.Background(), chainId)
		setup.UpdatePublicKey(context.Background(), chainId)
		time.Sleep(time.Second * 1)
		return
	}

	var wg sync.WaitGroup
	mutex := sync.Mutex{}

	var (
		numWorkers = 30
	)

	if *sweepBlockHeight < *cacheBlockHeight {
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				mutex.Lock()
				currentHeight := *sweepBlockHeight
				if currentHeight > *cacheBlockHeight {
					mutex.Unlock()
					return
				}
				*sweepBlockHeight++
				mutex.Unlock()

				err := SweepSolBlockchainTransactionCore(chainId, publicKey, sweepCount, currentHeight, constantSweepBlock, constantPendingBlock, constantPendingTransaction)
				if err != nil {
					_, err = global.NODE_REDIS.RPush(context.Background(), constantPendingBlock, currentHeight).Result()
					if err != nil {
						global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					}
				}
			}()
		}

		wg.Wait()

		_, err := global.NODE_REDIS.Set(context.Background(), constantSweepBlock, *sweepBlockHeight+1, 0).Result()
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return
		}
	}
}

func SweepSolBlockchainTransactionCore(
	chainId uint,
	publicKey *[]string,
	sweepCount *map[int64]int,
	sweepBlockHeight int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) error {
	defer utils.HandlePanic()

	var err error

	endpoint := constant.GetRPCUrlByNetwork(chainId)
	client := rpc.New(endpoint)

	includeRewards := false

	blockResult, err := client.GetBlockWithOpts(context.Background(), uint64(sweepBlockHeight), &rpc.GetBlockOpts{
		Encoding:                       solana.EncodingBase64,
		Commitment:                     rpc.CommitmentConfirmed,
		Rewards:                        &includeRewards,
		MaxSupportedTransactionVersion: &rpc.MaxSupportedTransactionVersion0,
	})

	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return err
	}

	if len(blockResult.Transactions) > 0 {
		for _, transaction := range blockResult.Transactions {

			isMonitorTx := false
			matchArray := make([]solana.PublicKey, 0)

			parsedTx, err := transaction.GetTransaction()
			if err != nil {
				return err
			}

			for _, inst := range parsedTx.Message.Instructions {
				programID := parsedTx.Message.AccountKeys[inst.ProgramIDIndex]

				switch programID {
				case solana.SystemProgramID:
					if len(inst.Data) >= 12 && binary.LittleEndian.Uint32(inst.Data[0:4]) == 2 {
						from := parsedTx.Message.AccountKeys[inst.Accounts[0]]
						to := parsedTx.Message.AccountKeys[inst.Accounts[1]]

						matchArray = append(matchArray, from, to)
					}

				case solana.TokenProgramID:
					if len(inst.Data) < 1 {
						break
					}

					var fromAccount, toAccount, mintAccount solana.PublicKey
					var fromTokenAccount, toTokenAccount token.Account

					switch inst.Data[0] {
					case 3:

						if len(inst.Data) < 9 || len(inst.Accounts) < 3 {
							break
						}

						fromAccount = parsedTx.Message.AccountKeys[inst.Accounts[0]]
						toAccount = parsedTx.Message.AccountKeys[inst.Accounts[1]]

						fromAccountInfo, err := client.GetAccountInfoWithOpts(context.Background(), fromAccount, &rpc.GetAccountInfoOpts{
							Encoding: solana.EncodingBase64,
						})
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							break
						}

						fromDecoder := bin.NewBinDecoder(fromAccountInfo.Value.Data.GetBinary())
						err = fromTokenAccount.UnmarshalWithDecoder(fromDecoder)
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							break
						}

						toAccountInfo, err := client.GetAccountInfoWithOpts(context.Background(), toAccount, &rpc.GetAccountInfoOpts{
							Encoding: solana.EncodingBase64,
						})
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							break
						}

						toDecoder := bin.NewBinDecoder(toAccountInfo.Value.Data.GetBinary())
						err = toTokenAccount.UnmarshalWithDecoder(toDecoder)
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							break
						}

						if !fromTokenAccount.Mint.Equals(toTokenAccount.Mint) {
							break
						}

						isSupportContract, _, _, _ := sweepUtils.GetContractInfo(chainId, fromTokenAccount.Mint.String())
						if !isSupportContract {
							break
						}

						matchArray = append(matchArray, fromTokenAccount.Owner, toTokenAccount.Owner)

					case 12:

						fromAccount = parsedTx.Message.AccountKeys[inst.Accounts[0]]
						mintAccount = parsedTx.Message.AccountKeys[inst.Accounts[1]]
						toAccount = parsedTx.Message.AccountKeys[inst.Accounts[2]]

						fromAccountInfo, err := client.GetAccountInfoWithOpts(context.Background(), fromAccount, &rpc.GetAccountInfoOpts{
							Encoding: solana.EncodingBase64,
						})
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							break
						}

						fromDecoder := bin.NewBinDecoder(fromAccountInfo.Value.Data.GetBinary())
						err = fromTokenAccount.UnmarshalWithDecoder(fromDecoder)
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							break
						}

						toAccountInfo, err := client.GetAccountInfoWithOpts(context.Background(), toAccount, &rpc.GetAccountInfoOpts{
							Encoding: solana.EncodingBase64,
						})
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							break
						}

						toDecoder := bin.NewBinDecoder(toAccountInfo.Value.Data.GetBinary())
						err = toTokenAccount.UnmarshalWithDecoder(toDecoder)
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							break
						}

						isSupportContract, _, _, _ := sweepUtils.GetContractInfo(chainId, mintAccount.String())
						if !isSupportContract {
							break
						}

						matchArray = append(matchArray, fromTokenAccount.Owner, toTokenAccount.Owner)
					default:
						break
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
						continue outerCurrentTxLoop
					}
				}
			}

			if isMonitorTx {
				redisTxs, err := global.NODE_REDIS.LRange(context.Background(), constantPendingTransaction, 0, -1).Result()
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					return err
				}

				sign := parsedTx.Signatures[0].String()

				for _, redisTx := range redisTxs {
					if redisTx == sign {
						break
					}
				}

				_, err = global.NODE_REDIS.RPush(context.Background(), constantPendingTransaction, sign).Result()
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					return err
				}
			}
		}
	}

	return nil
}

func SweepSolBlockchainTransactionDetails(
	chainId uint,
	publicKey *[]string,
	constantPendingTransaction string,
) {
	defer utils.HandlePanic()

	txHash, err := global.NODE_REDIS.LIndex(context.Background(), constantPendingTransaction, 0).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			time.Sleep(2 * time.Second)
			return
		}
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	global.NODE_LOG.Info(fmt.Sprintf("%s -> handle tx: %s", constant.GetChainName(chainId), txHash))

	endpoint := constant.GetRPCUrlByNetwork(chainId)
	client := rpc.New(endpoint)

	txSig := solana.MustSignatureFromBase58(txHash)

	version := uint64(0)
	transactionResult, err := client.GetTransaction(context.TODO(), txSig, &rpc.GetTransactionOpts{
		MaxSupportedTransactionVersion: &version,
		Encoding:                       solana.EncodingBase64,
	})
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	transaction, err := transactionResult.Transaction.GetTransaction()
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	var notifyRequest request.NotificationRequest

	notifyRequest.Hash = transaction.Signatures[0].String()
	notifyRequest.Chain = chainId

	if transactionResult.BlockTime != nil {
		notifyRequest.BlockTimestamp = int(transactionResult.BlockTime.Time().Unix()) * 1000
	}

	var isProcess bool

	for _, inst := range transaction.Message.Instructions {
		programID := transaction.Message.AccountKeys[inst.ProgramIDIndex]

		switch programID {
		case solana.SystemProgramID:
			if len(inst.Data) >= 12 && binary.LittleEndian.Uint32(inst.Data[0:4]) == 2 {
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

						err = notification.NotificationRequest(notifyRequest)
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							return
						}
						isProcess = true
					}

					if targetAddress.Equals(to) {
						notifyRequest.TransactType = "receive"
						notifyRequest.Address = v

						err = notification.NotificationRequest(notifyRequest)
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							return
						}
						isProcess = true
					}
				}

			}

		case solana.TokenProgramID:
			if len(inst.Data) < 1 {
				break
			}

			var amount uint64
			var fromAccount, toAccount, from, to, mintAccount solana.PublicKey
			var fromTokenAccount, toTokenAccount token.Account

			switch inst.Data[0] {
			case 3:

				if len(inst.Data) < 9 || len(inst.Accounts) < 3 {
					break
				}

				// source = transaction.Message.AccountKeys[inst.Accounts[0]]
				amount = binary.LittleEndian.Uint64(inst.Data[1:9])
				fromAccount = transaction.Message.AccountKeys[inst.Accounts[0]]
				toAccount = transaction.Message.AccountKeys[inst.Accounts[1]]
				// from = transaction.Message.AccountKeys[inst.Accounts[2]]

				fromAccountInfo, err := client.GetAccountInfoWithOpts(context.Background(), fromAccount, &rpc.GetAccountInfoOpts{
					Encoding: solana.EncodingBase64,
				})
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					return
				}

				fromDecoder := bin.NewBinDecoder(fromAccountInfo.Value.Data.GetBinary())
				err = fromTokenAccount.UnmarshalWithDecoder(fromDecoder)
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					return
				}

				toAccountInfo, err := client.GetAccountInfoWithOpts(context.Background(), toAccount, &rpc.GetAccountInfoOpts{
					Encoding: solana.EncodingBase64,
				})
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					return
				}

				toDecoder := bin.NewBinDecoder(toAccountInfo.Value.Data.GetBinary())
				err = toTokenAccount.UnmarshalWithDecoder(toDecoder)
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					return
				}

				if !fromTokenAccount.Mint.Equals(toTokenAccount.Mint) {
					return
				}

				isSupportContract, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, fromTokenAccount.Mint.String())
				if !isSupportContract {
					break
				}

				notifyRequest.FromAddress = fromTokenAccount.Owner.String()
				notifyRequest.ToAddress = toTokenAccount.Owner.String()
				notifyRequest.Token = contractName
				notifyRequest.Amount = utils.CalculateBalance(big.NewInt(int64(amount)), decimals)

				from = fromTokenAccount.Owner
				to = toTokenAccount.Owner

			case 12:

				amount = binary.LittleEndian.Uint64(inst.Data[1:9])
				fromAccount = transaction.Message.AccountKeys[inst.Accounts[0]]
				mintAccount = transaction.Message.AccountKeys[inst.Accounts[1]]
				toAccount = transaction.Message.AccountKeys[inst.Accounts[2]]
				// source = transaction.Message.AccountKeys[inst.Accounts[3]]

				fromAccountInfo, err := client.GetAccountInfoWithOpts(context.Background(), fromAccount, &rpc.GetAccountInfoOpts{
					Encoding: solana.EncodingBase64,
				})
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					return
				}

				fromDecoder := bin.NewBinDecoder(fromAccountInfo.Value.Data.GetBinary())
				err = fromTokenAccount.UnmarshalWithDecoder(fromDecoder)
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					return
				}

				toAccountInfo, err := client.GetAccountInfoWithOpts(context.Background(), toAccount, &rpc.GetAccountInfoOpts{
					Encoding: solana.EncodingBase64,
				})
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					return
				}

				toDecoder := bin.NewBinDecoder(toAccountInfo.Value.Data.GetBinary())
				err = toTokenAccount.UnmarshalWithDecoder(toDecoder)
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					return
				}

				isSupportContract, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, mintAccount.String())
				if !isSupportContract {
					break
				}

				notifyRequest.FromAddress = fromTokenAccount.Owner.String()
				notifyRequest.ToAddress = toTokenAccount.Owner.String()
				notifyRequest.Token = contractName
				notifyRequest.Amount = utils.CalculateBalance(big.NewInt(int64(amount)), decimals)

				from = fromTokenAccount.Owner
				to = toTokenAccount.Owner
			default:
				break
			}

			for _, v := range *publicKey {
				targetAddress := solana.MustPublicKeyFromBase58(v)

				if targetAddress.Equals(from) {
					notifyRequest.TransactType = "send"
					notifyRequest.Address = v

					err = notification.NotificationRequest(notifyRequest)
					if err != nil {
						global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
						return
					}
					isProcess = true
				}

				if targetAddress.Equals(to) {
					notifyRequest.TransactType = "receive"
					notifyRequest.Address = v

					err = notification.NotificationRequest(notifyRequest)
					if err != nil {
						global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
						return
					}
					isProcess = true
				}
			}
		default:
			continue
		}
	}

	if isProcess {
		_, err = global.NODE_REDIS.LPop(context.Background(), constantPendingTransaction).Result()
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		}
	} else {
		global.NODE_LOG.Error(fmt.Sprintf("Can not handle the tx: %s, Retry | %s -> %s", txHash, constant.GetChainName(chainId), err.Error()))
	}
}

func SweepSolBlockchainPendingBlock(
	chainId uint,
	publicKey *[]string,
	constantPendingBlock, constantPendingTransaction string,
) {
	defer utils.HandlePanic()

	blockHeight, err := global.NODE_REDIS.LIndex(context.Background(), constantPendingBlock, 0).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			time.Sleep(10 * time.Second)
			return
		}
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
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

	blockResult, err := client.GetBlockWithOpts(context.Background(), uint64(blockHeightInt), &rpc.GetBlockOpts{
		Encoding:                       solana.EncodingBase64,
		Commitment:                     rpc.CommitmentConfirmed,
		Rewards:                        &includeRewards,
		MaxSupportedTransactionVersion: &rpc.MaxSupportedTransactionVersion0,
	})

	if err != nil {
		// https://www.quicknode.com/docs/solana/error-references
		// 32007 Slot xxxxxx was skipped, or missing due to ledger jump to recent snapshot
		_, err = global.NODE_REDIS.LPop(context.Background(), constantPendingBlock).Result()
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		}
		return
	}

	if len(blockResult.Transactions) > 0 {
		for _, transaction := range blockResult.Transactions {

			isMonitorTx := false
			matchArray := make([]solana.PublicKey, 0)

			parsedTx, err := transaction.GetTransaction()
			if err != nil {
				return
			}

			for _, key := range parsedTx.Message.AccountKeys {
				matchArray = append(matchArray, key)
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
						continue outerCurrentTxLoop
					}
				}
			}

			if isMonitorTx {
				redisTxs, err := global.NODE_REDIS.LRange(context.Background(), constantPendingTransaction, 0, -1).Result()
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					return
				}

				sign := parsedTx.Signatures[0].String()

				for _, redisTx := range redisTxs {
					if redisTx == sign {
						break
					}
				}

				_, err = global.NODE_REDIS.RPush(context.Background(), constantPendingTransaction, sign).Result()
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					return
				}
			}
		}
	}

	_, err = global.NODE_REDIS.LPop(context.Background(), constantPendingBlock).Result()
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
	}
}
