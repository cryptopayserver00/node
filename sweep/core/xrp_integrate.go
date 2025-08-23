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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xrpscan/xrpl-go"

	sweepUtils "node/sweep/utils"
)

func SetupXrpLatestBlockHeight(chainId uint) {
	config := xrpl.ClientConfig{
		URL: constant.XrpWsByNetwork(chainId),
	}
	client := xrpl.NewClient(config)
	err := client.Ping([]byte("PING"))
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	request := xrpl.BaseRequest{
		"id":      2,
		"command": "ledger_current",
	}

	xrpResponse, err := client.Request(request)
	if err == nil && xrpResponse["status"] == "success" {
		result, ok := xrpResponse["result"].(map[string]any)
		if !ok {
			return
		}

		ledgerIndex, ok := result["ledger_current_index"].(float64)
		if !ok {
			return
		}

		if ledgerIndex > 0 {
			setup.SetupLatestBlockHeight(context.Background(), chainId, int64(ledgerIndex))
		}
	}
}

func SweepXrpBlockchainTransaction(
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepCount *map[int64]int,
	sweepBlockHeight, cacheBlockHeight *int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) {
	defer utils.HandlePanic()

	if len(*publicKey) <= 0 {
		SetupXrpLatestBlockHeight(chainId)
		setup.UpdateCacheBlockHeight(context.Background(), chainId)
		setup.UpdateSweepBlockHeight(context.Background(), chainId)
		setup.UpdatePublicKey(context.Background(), chainId)
		return
	}

	if *sweepBlockHeight >= *cacheBlockHeight {
		SetupXrpLatestBlockHeight(chainId)
		setup.UpdateCacheBlockHeight(context.Background(), chainId)
		setup.UpdatePublicKey(context.Background(), chainId)
		time.Sleep(time.Second * 1)
		return
	}

	var wg sync.WaitGroup
	mutex := sync.Mutex{}

	var (
		numWorkers = 10
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

				err := SweepXrpBlockchainTransactionCore(client, chainId, publicKey, sweepCount, currentHeight, constantSweepBlock, constantPendingBlock, constantPendingTransaction)
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

func SweepXrpBlockchainTransactionCore(
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepCount *map[int64]int,
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

	err = client.HTTPPost(jsonRpcRequest, &rpcBlockDetail)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return err
	}

	if sweepBlockHeight == int64(rpcBlockDetail.Result.LedgerIndex) {
		if len(rpcBlockDetail.Result.Ledger.Transactions) > 0 {
			for _, transaction := range rpcBlockDetail.Result.Ledger.Transactions {

				isMonitorTx := false
				txFrom := transaction.TxJson.Account
				txTo := transaction.TxJson.Destination

				matchArray := make([]string, 0)

				// xrp
				_, ok := transaction.TxJson.DeliverMax.(string)
				if ok {
					matchArray = append(matchArray, txFrom, txTo)
				}

				// token
				tokenResult, ok := transaction.TxJson.DeliverMax.(map[string]any)
				if ok {
					if isSupportContract, _, _, _ := sweepUtils.GetContractInfo(chainId, tokenResult["issuer"].(string)); isSupportContract {
						matchArray = append(matchArray, txFrom, txTo)
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

					for _, redisTx := range redisTxs {
						if redisTx == transaction.Hash {
							break
						}
					}

					_, err = global.NODE_REDIS.RPush(context.Background(), constantPendingTransaction, transaction.Hash).Result()
					if err != nil {
						global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
						return err
					}
				}
			}
		}

		return nil
	} else {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", sweepBlockHeight, rpcBlockDetail.Result.LedgerIndex)))
		return errors.New("not the same height of block")
	}
}

func SweepXrpBlockchainTransactionDetails(
	client NODE_Client.Client,
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

	err = client.HTTPPost(jsonRpcRequest, &rpcTransactionDetail)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	var notifyRequest request.NotificationRequest

	notifyRequest.Hash = rpcTransactionDetail.Result.Hash
	notifyRequest.Chain = chainId

	timestamp, err := time.Parse(time.RFC3339, rpcTransactionDetail.Result.CloseTimeIso)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}
	notifyRequest.BlockTimestamp = int(timestamp.Unix()) * 1000

	var isProcess = false

	notifyRequest.FromAddress = rpcTransactionDetail.Result.TxJson.Account
	notifyRequest.ToAddress = rpcTransactionDetail.Result.TxJson.Destination

	// xrp
	xrpResult, ok := rpcTransactionDetail.Result.TxJson.DeliverMax.(string)
	if ok {
		isSupportContract, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, "")
		if isSupportContract {
			notifyRequest.Token = contractName
			amount, err := strconv.ParseInt(xrpResult, 10, 64)
			if err != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				return
			}
			notifyRequest.Amount = utils.CalculateBalance(big.NewInt(amount), decimals)
		}
	}

	// token
	tokenResult, ok := rpcTransactionDetail.Result.TxJson.DeliverMax.(map[string]any)
	if ok {
		if isSupportContract, contractName, _, _ := sweepUtils.GetContractInfo(chainId, tokenResult["issuer"].(string)); isSupportContract {
			notifyRequest.Token = contractName
			notifyRequest.Amount = tokenResult["value"].(string)
		}
	}

	for _, v := range *publicKey {
		if strings.EqualFold(v, notifyRequest.FromAddress) {
			notifyRequest.TransactType = "send"
			notifyRequest.Address = v

			err = notification.NotificationRequest(notifyRequest)
			if err != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				return
			}
			isProcess = true
		}

		if strings.EqualFold(v, notifyRequest.ToAddress) {
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

	if !isProcess || err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("Can not handle the tx: %s, Retry | %s -> %s", txHash, constant.GetChainName(chainId), err.Error()))
		return
	}

	_, err = global.NODE_REDIS.LPop(context.Background(), constantPendingTransaction).Result()
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}
}

func SweepXrpBlockchainPendingBlock(
	client NODE_Client.Client,
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

	err = client.HTTPPost(jsonRpcRequest, &rpcBlockDetail)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	if blockHeightInt == int64(rpcBlockDetail.Result.LedgerIndex) {
		if len(rpcBlockDetail.Result.Ledger.Transactions) > 0 {
			for _, transaction := range rpcBlockDetail.Result.Ledger.Transactions {

				isMonitorTx := false
				txFrom := transaction.TxJson.Account
				txTo := transaction.TxJson.Destination

				matchArray := make([]string, 0)

				// xrp
				_, ok := transaction.TxJson.DeliverMax.(string)
				if ok {
					matchArray = append(matchArray, txFrom, txTo)
				}

				// token
				tokenResult, ok := transaction.TxJson.DeliverMax.(map[string]any)
				if ok {
					if isSupportContract, _, _, _ := sweepUtils.GetContractInfo(chainId, tokenResult["issuer"].(string)); isSupportContract {
						matchArray = append(matchArray, txFrom, txTo)
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

					for _, redisTx := range redisTxs {
						if redisTx == transaction.Hash {
							break
						}
					}

					_, err = global.NODE_REDIS.RPush(context.Background(), constantPendingTransaction, transaction.Hash).Result()
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
	} else {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", blockHeightInt, rpcBlockDetail.Result.LedgerIndex)))
	}
}
