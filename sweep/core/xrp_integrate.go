package core

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
	"node/sweep/setup"
	"node/utils"
	NODE_Client "node/utils/http"
	"node/utils/notification"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	sweepUtils "node/sweep/utils"

	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/redis/go-redis/v9"
)

func SetupXrpLatestBlockHeight(ctx context.Context, chainId uint) {
	config, err := rpc.NewClientConfig(
		constant.XrpWsByNetwork(chainId),
	)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return
	}

	client := rpc.NewClient(config)
	ledgerIndex, err := client.GetLedgerIndex()
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return
	}

	if int64(ledgerIndex) > 0 {
		setup.SetupLatestBlockHeight(ctx, chainId, int64(ledgerIndex))

	}
}

func SweepXrpBlockchainTransaction(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepBlockHeight, cacheBlockHeight *int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) {
	defer utils.HandlePanic()

	if len(*publicKey) <= 0 {
		SetupXrpLatestBlockHeight(ctx, chainId)
		setup.UpdateCacheBlockHeight(ctx, chainId)
		setup.UpdateSweepBlockHeight(ctx, chainId)
		setup.UpdatePublicKey(ctx, chainId)
		return
	}

	if *sweepBlockHeight >= *cacheBlockHeight {
		SetupXrpLatestBlockHeight(ctx, chainId)
		setup.UpdateCacheBlockHeight(ctx, chainId)
		setup.UpdatePublicKey(ctx, chainId)
		time.Sleep(time.Second * 5)
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

			err := SweepXrpBlockchainTransactionCore(ctx, client, chainId, publicKey, height, constantSweepBlock, constantPendingBlock, constantPendingTransaction)
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

func SweepXrpBlockchainTransactionCore(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepBlockHeight int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) error {
	defer utils.HandlePanic()

	var err error

	client.URL = constant.GetRPCUrlByNetwork(chainId)
	var rpcBlockDetail response.XrpscanBlockResponse
	var jsonRpcRequest request.XrpJsonRpcRequest
	jsonRpcRequest.Method = "ledger"
	jsonRpcRequest.Params = []map[string]any{
		{
			"ledger_index": sweepBlockHeight,
			"transactions": true,
			"expand":       true,
			"api_version":  2,
		},
	}

	err = client.HTTPPost(ctx, jsonRpcRequest, &rpcBlockDetail)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	if rpcBlockDetail.Result.LedgerIndex == 0 {
		return errors.New("not support")
	}

	if sweepBlockHeight == int64(rpcBlockDetail.Result.LedgerIndex) {
		if len(rpcBlockDetail.Result.Ledger.Transactions) > 0 {
			for _, transaction := range rpcBlockDetail.Result.Ledger.Transactions {

				if transaction.Meta.TransactionResult != "tesSUCCESS" {
					continue
				}

				isMonitorTx := false
				txFrom := transaction.TxJson.Account
				txTo := transaction.TxJson.Destination

				matchArray := make([]string, 0)

				deliveredResult := transaction.Meta.DeliveredAmount

				if _, ok := deliveredResult.(string); ok {
					matchArray = append(matchArray, txFrom, txTo)
				} else if tokenResult, ok := deliveredResult.(map[string]any); ok {
					issuer, issuerOk := tokenResult["issuer"].(string)
					if issuerOk {
						if isSupportContract, _, _, _ := sweepUtils.GetContractInfo(chainId, issuer); isSupportContract {
							matchArray = append(matchArray, txFrom, txTo)
						}
					}
				}

				if len(matchArray) == 0 {
					continue
				}

				matchArray = utils.RemoveDuplicatesForString(matchArray)

			outerCurrentTxLoop:
				for i := 0; i < len(*publicKey); i++ {
					for _, j := range matchArray {
						if strings.EqualFold((*publicKey)[i], j) {
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

					found := slices.Contains(redisTxs, transaction.Hash)
					if !found {
						_, err = global.NODE_REDIS.RPush(ctx, constantPendingTransaction, transaction.Hash).Result()
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
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", sweepBlockHeight, rpcBlockDetail.Result.LedgerIndex)))
		time.Sleep(2 * time.Second)
		return errors.New("not the same height of block")
	}
}

func SweepXrpBlockchainTransactionDetails(
	ctx context.Context,
	client NODE_Client.Client,
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

	err = handleXrpTransactionDetails(ctx, client, chainId, publicKey, txHash)

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

func handleXrpTransactionDetails(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	txHash string,
) error {
	client.URL = constant.GetRPCUrlByNetwork(chainId)
	var rpcTransactionDetail response.XrpscanTransactionResponse
	var jsonRpcRequest request.XrpJsonRpcRequest
	jsonRpcRequest.Method = "tx"
	jsonRpcRequest.Params = []map[string]any{
		{
			"transaction": txHash,
			"binary":      false,
			"api_version": 2,
		},
	}

	err := client.HTTPPost(ctx, jsonRpcRequest, &rpcTransactionDetail)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	if rpcTransactionDetail.Result.Meta.TransactionResult != "tesSUCCESS" {
		// 交易失败，直接跳过（不重试，因为它不是"处理出错"，是链上本身执行失败）
		return errors.New("not support")
	}

	var notifyRequest request.NotificationRequest

	notifyRequest.Hash = rpcTransactionDetail.Result.Hash
	notifyRequest.Chain = chainId

	timestamp, err := time.Parse(time.RFC3339, rpcTransactionDetail.Result.CloseTimeIso)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}
	notifyRequest.BlockTimestamp = int(timestamp.Unix()) * 1000
	notifyRequest.FromAddress = rpcTransactionDetail.Result.TxJson.Account
	notifyRequest.ToAddress = rpcTransactionDetail.Result.TxJson.Destination

	// 优先使用 meta.delivered_amount（真实到账金额），而不是 tx_json.DeliverMax（意图金额上限）
	deliveredResult := rpcTransactionDetail.Result.Meta.DeliveredAmount

	if xrpResult, ok := deliveredResult.(string); ok {
		isSupportContract, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, "")
		if !isSupportContract {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> unsupported xrp, hash: %s, dropping", constant.GetChainName(chainId), notifyRequest.Hash))
			return errors.New("not support")
		} else {
			notifyRequest.Token = contractName
			amount, err := strconv.ParseInt(xrpResult, 10, 64)
			if err != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				time.Sleep(2 * time.Second)
				return err
			}
			notifyRequest.Amount = utils.CalculateBalance(big.NewInt(amount), decimals)
		}
	} else if tokenResult, ok := deliveredResult.(map[string]any); ok {
		issuer, issuerOk := tokenResult["issuer"].(string)
		value, valueOk := tokenResult["value"].(string)
		if !issuerOk || !valueOk {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> malformed token DeliverMax, hash: %s, dropping", constant.GetChainName(chainId), notifyRequest.Hash))
			return errors.New("not support")
		}

		isSupportContract, contractName, _, _ := sweepUtils.GetContractInfo(chainId, issuer)
		if !isSupportContract {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> unsupported token issuer: %s, hash: %s, dropping", constant.GetChainName(chainId), issuer, notifyRequest.Hash))
			return errors.New("not support")
		} else {
			notifyRequest.Token = contractName
			notifyRequest.Amount = value
		}
	} else {
		return errors.New("not support")
	}

	isProcess := false

	for _, v := range *publicKey {
		if strings.EqualFold(v, notifyRequest.FromAddress) {
			notifyRequest.TransactType = "send"
			notifyRequest.Address = v

			err = notification.NotificationRequest(ctx, notifyRequest)
			if err != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				return err
			}
			isProcess = true
		}

		if strings.EqualFold(v, notifyRequest.ToAddress) {
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

	if !isProcess {
		return fmt.Errorf("no monitored address matched for tx: %s", notifyRequest.Hash)
	}

	return nil
}

func SweepXrpBlockchainPendingBlock(
	ctx context.Context,
	client NODE_Client.Client,
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

	client.URL = constant.GetRPCUrlByNetwork(chainId)
	var rpcBlockDetail response.XrpscanBlockResponse
	var jsonRpcRequest request.XrpJsonRpcRequest
	jsonRpcRequest.Method = "ledger"
	jsonRpcRequest.Params = []map[string]any{
		{
			"ledger_index": blockHeightInt,
			"transactions": true,
			"expand":       true,
			"api_version":  2,
		},
	}

	err = client.HTTPPost(ctx, jsonRpcRequest, &rpcBlockDetail)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return
	}

	if rpcBlockDetail.Result.LedgerIndex == 0 {
		return
	}

	if blockHeightInt == int64(rpcBlockDetail.Result.LedgerIndex) {
		if len(rpcBlockDetail.Result.Ledger.Transactions) > 0 {
			for _, transaction := range rpcBlockDetail.Result.Ledger.Transactions {

				if transaction.Meta.TransactionResult != "tesSUCCESS" {
					continue
				}

				isMonitorTx := false
				txFrom := transaction.TxJson.Account
				txTo := transaction.TxJson.Destination

				matchArray := make([]string, 0)

				deliveredResult := transaction.Meta.DeliveredAmount

				if _, ok := deliveredResult.(string); ok {
					matchArray = append(matchArray, txFrom, txTo)
				} else if tokenResult, ok := deliveredResult.(map[string]any); ok {
					issuer, issuerOk := tokenResult["issuer"].(string)
					if issuerOk {
						if isSupportContract, _, _, _ := sweepUtils.GetContractInfo(chainId, issuer); isSupportContract {
							matchArray = append(matchArray, txFrom, txTo)
						}
					}
				}

				if len(matchArray) == 0 {
					continue
				}

				matchArray = utils.RemoveDuplicatesForString(matchArray)

			outerCurrentTxLoop:
				for i := 0; i < len(*publicKey); i++ {
					for _, j := range matchArray {
						if strings.EqualFold((*publicKey)[i], j) {
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

					found := slices.Contains(redisTxs, transaction.Hash)
					if !found {
						_, err = global.NODE_REDIS.RPush(ctx, constantPendingTransaction, transaction.Hash).Result()
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
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", blockHeightInt, rpcBlockDetail.Result.LedgerIndex)))

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
