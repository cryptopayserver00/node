package core

import (
	"context"
	"errors"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
	"node/sweep/setup"
	"node/sweep/utils/erc20"
	"node/utils"
	"node/utils/redisq"
	"strconv"
	"sync"
	"time"

	sweepUtils "node/sweep/utils"
	NODE_Client "node/utils/http"
	"node/utils/notification"
)

const maxWorkers = 10

func calcNumWorkers(sweepBlockHeight, cacheBlockHeight int64, maxWorkers int) int {
	remaining := cacheBlockHeight - sweepBlockHeight + 1
	if remaining <= 0 {
		return 0
	}
	if remaining < int64(maxWorkers) {
		return int(remaining)
	}
	return maxWorkers
}

func SetupLatestBlockHeight(ctx context.Context, client NODE_Client.Client, chainId uint) {
	var err error
	client.URL = constant.GetRPCUrlByChainId(chainId)
	var rpcBlockInfo response.RPCBlockInfo
	var jsonRpcRequest request.JsonRpcRequest
	jsonRpcRequest.Id = 1
	jsonRpcRequest.Jsonrpc = "2.0"
	jsonRpcRequest.Method = "eth_getBlockByNumber"
	jsonRpcRequest.Params = []any{"latest", false}

	err = client.HTTPPost(ctx, jsonRpcRequest, &rpcBlockInfo)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return
	}

	if rpcBlockInfo.Result.Number == "" {
		return
	}

	height, err := utils.HexStringToUint64(rpcBlockInfo.Result.Number)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	if height > 0 {
		setup.SetupLatestBlockHeight(ctx, chainId, int64(height))
	}
}

func SweepBlockchainTransaction(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepBlockHeight, cacheBlockHeight *int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) {
	defer utils.HandlePanic()

	if len(*publicKey) <= 0 {
		SetupLatestBlockHeight(ctx, client, chainId)
		setup.UpdateCacheBlockHeight(ctx, chainId)
		setup.UpdateSweepBlockHeight(ctx, chainId)
		setup.UpdatePublicKey(ctx, chainId)
		return
	}

	if *sweepBlockHeight >= *cacheBlockHeight {
		SetupLatestBlockHeight(ctx, client, chainId)
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

			var err error
			if global.NODE_CONFIG.Blockchain.InnerTx {
				err = SweepBlockchainTransactionCoreForEthereum(ctx, client, chainId, publicKey, height, constantSweepBlock, constantPendingBlock, constantPendingTransaction)
			} else {
				err = SweepBlockchainTransactionCore(ctx, client, chainId, publicKey, height, constantSweepBlock, constantPendingBlock, constantPendingTransaction)
			}
			if err != nil {
				if _, rpushErr := redisq.EnqueueUnique(ctx, constantPendingBlock, strconv.FormatInt(height, 10)); rpushErr != nil {
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

func SweepBlockchainTransactionCore(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepBlockHeight int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) error {
	defer utils.HandlePanic()

	var err error

	client.URL = constant.GetRPCUrlByChainId(chainId)
	var rpcBlockDetail response.RPCBlockDetail
	var jsonRpcRequest request.JsonRpcRequest
	jsonRpcRequest.Id = 1
	jsonRpcRequest.Jsonrpc = "2.0"
	jsonRpcRequest.Method = "eth_getBlockByNumber"
	jsonRpcRequest.Params = []any{"0x" + strconv.FormatInt(sweepBlockHeight, 16), true}

	err = client.HTTPPost(ctx, jsonRpcRequest, &rpcBlockDetail)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	if rpcBlockDetail.Result.Number == "" {
		err = fmt.Errorf("can not get the number: %s", rpcBlockDetail.Result.Number)
		time.Sleep(2 * time.Second)
		return err
	}

	height, err := utils.HexStringToUint64(rpcBlockDetail.Result.Number)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	if sweepBlockHeight == int64(height) {
		monitorSet := utils.BuildHexAddressSet(*publicKey)
		if len(rpcBlockDetail.Result.Transactions) > 0 {
			for _, transaction := range rpcBlockDetail.Result.Transactions {

				isMonitorTx := false
				txFrom := utils.HexToAddress(transaction.From)
				txTo := utils.HexToAddress(transaction.To)

				matchArray := make([]string, 0)

				if transaction.Input == "0x" {
					matchArray = append(matchArray, txFrom, txTo)
				} else {
					if isSupportContract, contractName, _, _ := sweepUtils.GetContractInfo(chainId, txTo); isSupportContract {
						if arrays, err := erc20.GetAllAddressByTransactionTwo(chainId, contractName, txFrom, transaction.Hash, transaction.Input); err == nil {
							matchArray = append(matchArray, arrays...)
						}
					}
				}

				matchArray = utils.RemoveDuplicatesForString(matchArray)

				if len(matchArray) == 0 {
					continue
				}

			isMonitorTx = utils.MatchHexAddressSet(monitorSet, matchArray...)

				if isMonitorTx {
					var rpcReceipt response.RPCReceiptTransactionDetail
					jsonRpcRequest.Id = 1
					jsonRpcRequest.Jsonrpc = "2.0"
					jsonRpcRequest.Method = "eth_getTransactionReceipt"
					jsonRpcRequest.Params = []any{transaction.Hash}

					err = client.HTTPPost(ctx, jsonRpcRequest, &rpcReceipt)
					if err != nil {
						global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
						time.Sleep(2 * time.Second)
						return err
					}

					if rpcReceipt.Result.Status == "0x0" {
						continue
					}

					if _, err = redisq.EnqueueUnique(ctx, constantPendingTransaction, transaction.Hash); err != nil {
						global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
						time.Sleep(2 * time.Second)
						return err
					}
				}
			}
		}

		return nil
	} else {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", sweepBlockHeight, height)))
		time.Sleep(2 * time.Second)
		return errors.New("not the same height of block")
	}
}

func SweepBlockchainTransactionDetails(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	txRetryCount *map[string]int,
	constantPendingTransaction string,
) {
	defer utils.HandlePanic()

	txHash, err := redisq.Pop(ctx, constantPendingTransaction)
	if err != nil {
		if redisq.IsEmpty(err) {
			time.Sleep(2 * time.Second)
			return
		}
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return
	}

	global.NODE_LOG.Info(fmt.Sprintf("%s -> handle tx: %s", constant.GetChainName(chainId), txHash))

	err = handleTransactionDetails(ctx, client, chainId, publicKey, txHash)

	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("Can not handle the tx: %s, Retry | %s -> %s", txHash, constant.GetChainName(chainId), err.Error()))

		retryCount := (*txRetryCount)[txHash]
		retryCount++
		if retryCount >= setup.SweepThreshold {
			global.NODE_LOG.Error(fmt.Sprintf("give up tx after retries: %s", txHash))
			if ackErr := redisq.Ack(ctx, constantPendingTransaction, txHash); ackErr != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), ackErr.Error()))
			}
			delete(*txRetryCount, txHash)
			return
		}
		(*txRetryCount)[txHash] = retryCount
		if rqErr := redisq.Requeue(ctx, constantPendingTransaction, txHash); rqErr != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), rqErr.Error()))
		}
		time.Sleep(2 * time.Second)
		return
	}

	if ackErr := redisq.Ack(ctx, constantPendingTransaction, txHash); ackErr != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), ackErr.Error()))
		time.Sleep(2 * time.Second)
		return
	}
	delete(*txRetryCount, txHash)
}

func handleTransactionDetails(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	txHash string,
) error {
	var err error

	client.URL = constant.GetRPCUrlByChainId(chainId)
	var rpcDetail response.RPCTransactionDetail
	var jsonRpcRequest request.JsonRpcRequest
	jsonRpcRequest.Id = 1
	jsonRpcRequest.Jsonrpc = "2.0"
	jsonRpcRequest.Method = "eth_getTransactionByHash"
	jsonRpcRequest.Params = []any{txHash}

	err = client.HTTPPost(ctx, jsonRpcRequest, &rpcDetail)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	var rpcBlockInfo response.RPCBlockInfo
	jsonRpcRequest.Id = 1
	jsonRpcRequest.Jsonrpc = "2.0"
	jsonRpcRequest.Method = "eth_getBlockByNumber"
	jsonRpcRequest.Params = []any{rpcDetail.Result.BlockNumber, false}

	err = client.HTTPPost(ctx, jsonRpcRequest, &rpcBlockInfo)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	if rpcBlockInfo.Result.Timestamp == "" {
		return errors.New("not support")
	}

	blockTimeStamp, err := utils.HexStringToUint64(rpcBlockInfo.Result.Timestamp)
	if err != nil {
		return err
	}

	var notifyRequest request.NotificationRequest

	notifyRequest.Hash = rpcDetail.Result.Hash
	notifyRequest.Chain = chainId
	notifyRequest.BlockTimestamp = int(blockTimeStamp) * 1000

	if global.NODE_CONFIG.Blockchain.InnerTx {
		err = handleInternalTransaction(ctx, client, chainId, publicKey, notifyRequest, rpcDetail)
	} else {
		err = handleOrdinaryTransaction(ctx, chainId, publicKey, notifyRequest, rpcDetail.Result.From, rpcDetail.Result.To, rpcDetail.Result.Hash, rpcDetail.Result.Input, rpcDetail.Result.Value)
	}

	return err
}

func handleOrdinaryTransaction(ctx context.Context, chainId uint, publicKey *[]string, notifyRequest request.NotificationRequest, from, to, hash, input, value string) error {
	isProcess := false
	monitorSet := utils.BuildHexAddressSet(*publicKey)

	var (
		isSupportContract bool
		contractName      string
		decimals          int
	)

	if input == "0x" {
		isSupportContract, contractName, _, decimals = sweepUtils.GetContractInfo(chainId, constant.ETH_NATIVE_PLACEHOLDER_ADDRESS)
	} else {
		isSupportContract, contractName, _, decimals = sweepUtils.GetContractInfo(chainId, to)
	}

	if !isSupportContract {
		return errors.New("can not find the contract: " + to)
	}

	fromAddress := from
	notifyRequest.FromAddress = fromAddress

	if decimals == 0 {
		return errors.New("decimals can not be 0")
	}

	if !(input == "0x") {
		methodName, decodeFromAddress, decodeToAddress, transactionValue, err := erc20.DecodeERC20TransactionInputData(chainId, hash, input)
		if err != nil {
			return err
		}

		switch methodName {
		case erc20.TransferFrom:
			fromAddress = decodeFromAddress
			notifyRequest.FromAddress = fromAddress
		}

		notifyRequest.ToAddress = decodeToAddress
		notifyRequest.Token = contractName
		notifyRequest.Amount = utils.CalculateBalance(transactionValue, decimals)

		fromNorm := utils.HexToAddress(fromAddress)
		toNorm := utils.HexToAddress(decodeToAddress)
		if _, ok := monitorSet[fromNorm]; ok {
			notifyRequest.TransactType = "send"
			notifyRequest.Address = fromNorm
			if err = notification.NotificationRequest(ctx, notifyRequest); err != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				return err
			}
			isProcess = true
		}
		if _, ok := monitorSet[toNorm]; ok {
			notifyRequest.TransactType = "receive"
			notifyRequest.Address = toNorm
			if err = notification.NotificationRequest(ctx, notifyRequest); err != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				return err
			}
			isProcess = true
		}

	} else {
		toAddress := to
		notifyRequest.ToAddress = toAddress

		value, err := utils.HexStringToBigInt(value)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return err
		}
		notifyRequest.Amount = utils.CalculateBalance(value, decimals)
		notifyRequest.Token = contractName

		fromNorm := utils.HexToAddress(fromAddress)
		toNorm := utils.HexToAddress(toAddress)
		if _, ok := monitorSet[fromNorm]; ok {
			notifyRequest.TransactType = "send"
			notifyRequest.Address = fromNorm
			if err = notification.NotificationRequest(ctx, notifyRequest); err != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				return err
			}
			isProcess = true
		}
		if _, ok := monitorSet[toNorm]; ok {
			notifyRequest.TransactType = "receive"
			notifyRequest.Address = toNorm
			if err = notification.NotificationRequest(ctx, notifyRequest); err != nil {
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

func SweepBlockchainPendingBlock(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	retryBlockCount *map[int64]int,
	constantPendingBlock, constantPendingTransaction string,
) {
	defer utils.HandlePanic()

	blockHeight, err := redisq.Pop(ctx, constantPendingBlock)
	if err != nil {
		if redisq.IsEmpty(err) {
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
		_ = redisq.Ack(ctx, constantPendingBlock, blockHeight)
		return
	}

	monitorSet := utils.BuildHexAddressSet(*publicKey)
	finished := false
	defer func() {
		if !finished {
			if rqErr := redisq.Requeue(ctx, constantPendingBlock, blockHeight); rqErr != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> requeue pending block: %s", constant.GetChainName(chainId), rqErr.Error()))
			}
		}
	}()

	if global.NODE_CONFIG.Blockchain.InnerTx {
		var rpcBlockDetail response.RPCBlockInnerDetail
		client.URL = constant.GetAlchemyRPCUrlByChainId(chainId)
		payload := map[string]any{
			"id":      1,
			"jsonrpc": "2.0",
			"method":  "debug_traceBlockByNumber",
			"params": []any{
				"0x" + strconv.FormatInt(blockHeightInt, 16),
				map[string]any{
					"tracer": "callTracer",
				},
			},
		}

		err = client.HTTPPost(ctx, payload, &rpcBlockDetail)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			time.Sleep(2 * time.Second)
			return
		}

		if len(rpcBlockDetail.Result) == 0 {
			blockN, ok := (*retryBlockCount)[blockHeightInt]
			if !ok {
				(*retryBlockCount)[blockHeightInt] = 1

				err = errors.New("can not get the transaction of block number: " + fmt.Sprint(blockHeightInt))
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				return
			} else if blockN >= setup.SweepThreshold {
				delete(*retryBlockCount, blockHeightInt)
			} else {
				(*retryBlockCount)[blockHeightInt]++

				err = errors.New("can not get the transaction of block number: " + fmt.Sprint(blockHeightInt))
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				return
			}
		}

		if len(rpcBlockDetail.Result) > 0 {
			for _, transaction := range rpcBlockDetail.Result {
				isMonitorTx := false
				txFrom := utils.HexToAddress(transaction.Result.From)
				txTo := utils.HexToAddress(transaction.Result.To)

				matchArray := make([]string, 0)

				if transaction.Result.Error != "" {
					continue
				}

				if transaction.Result.Type == "CALL" {
					if transaction.Result.Input == "0x" {
						matchArray = append(matchArray, txFrom, txTo)
					} else {
						if isSupportContract, _, _, _ := sweepUtils.GetContractInfo(chainId, txTo); isSupportContract {
							if arrays, err := erc20.GetAllAddressByTransaction(chainId, txFrom, transaction.Hash, transaction.Result.Input); err == nil {
								matchArray = append(matchArray, arrays...)
							}
						}
					}
				}

				if len(transaction.Result.Calls) > 0 {
					matchArray = append(matchArray, processCallsForScanBlock(chainId, transaction.Hash, transaction.Result.Calls)...)
				}

				matchArray = utils.RemoveDuplicatesForString(matchArray)

				if len(matchArray) == 0 {
					continue
				}

			isMonitorTx = utils.MatchHexAddressSet(monitorSet, matchArray...)

				if isMonitorTx {
					var rpcReceipt response.RPCReceiptTransactionDetail
					var jsonRpcRequest request.JsonRpcRequest
					jsonRpcRequest.Id = 1
					jsonRpcRequest.Jsonrpc = "2.0"
					jsonRpcRequest.Method = "eth_getTransactionReceipt"
					jsonRpcRequest.Params = []any{transaction.Hash}

					err = client.HTTPPost(ctx, jsonRpcRequest, &rpcReceipt)
					if err != nil {
						global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
						time.Sleep(2 * time.Second)
						return
					}

					if rpcReceipt.Result.Status == "0x0" {
						continue
					}

					if _, err = redisq.EnqueueUnique(ctx, constantPendingTransaction, transaction.Hash); err != nil {
						global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
						time.Sleep(2 * time.Second)
						return
					}
				}
			}
		}

		if ackErr := redisq.Ack(ctx, constantPendingBlock, blockHeight); ackErr != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), ackErr.Error()))
			time.Sleep(2 * time.Second)
		}
		finished = true

		*retryBlockCount = make(map[int64]int)
	} else {
		client.URL = constant.GetRPCUrlByChainId(chainId)
		var rpcBlockDetail response.RPCBlockDetail
		var jsonRpcRequest request.JsonRpcRequest
		jsonRpcRequest.Id = 1
		jsonRpcRequest.Jsonrpc = "2.0"
		jsonRpcRequest.Method = "eth_getBlockByNumber"
		jsonRpcRequest.Params = []any{"0x" + strconv.FormatInt(blockHeightInt, 16), true}

		err = client.HTTPPost(ctx, jsonRpcRequest, &rpcBlockDetail)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			time.Sleep(2 * time.Second)
			return
		}

		if rpcBlockDetail.Result.Number == "" {
			err = fmt.Errorf("can not get the number: %s, %s", blockHeight, client.URL)
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			time.Sleep(2 * time.Second)
			return
		}

		height, err := utils.HexStringToUint64(rpcBlockDetail.Result.Number)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			time.Sleep(2 * time.Second)
			return
		}

		if blockHeightInt == int64(height) {
			if len(rpcBlockDetail.Result.Transactions) > 0 {
				for _, transaction := range rpcBlockDetail.Result.Transactions {

					isMonitorTx := false
					txFrom := utils.HexToAddress(transaction.From)
					txTo := utils.HexToAddress(transaction.To)

					matchArray := make([]string, 0)

					if transaction.Input == "0x" {
						matchArray = append(matchArray, txFrom, txTo)
					} else {
						if isSupportContract, contractName, _, _ := sweepUtils.GetContractInfo(chainId, txTo); isSupportContract {
							if arrays, err := erc20.GetAllAddressByTransactionTwo(chainId, contractName, txFrom, transaction.Hash, transaction.Input); err == nil {
								matchArray = append(matchArray, arrays...)
							}
						}
					}

					matchArray = utils.RemoveDuplicatesForString(matchArray)

					if len(matchArray) == 0 {
						continue
					}

				isMonitorTx = utils.MatchHexAddressSet(monitorSet, matchArray...)

					if isMonitorTx {
						var rpcReceipt response.RPCReceiptTransactionDetail
						jsonRpcRequest.Id = 1
						jsonRpcRequest.Jsonrpc = "2.0"
						jsonRpcRequest.Method = "eth_getTransactionReceipt"
						jsonRpcRequest.Params = []any{transaction.Hash}

						err = client.HTTPPost(ctx, jsonRpcRequest, &rpcReceipt)
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							time.Sleep(2 * time.Second)
							return
						}

						if rpcReceipt.Result.Status == "0x0" {
							continue
						}

						if _, err = redisq.EnqueueUnique(ctx, constantPendingTransaction, transaction.Hash); err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							time.Sleep(2 * time.Second)
							return
						}
					}
				}
			}

			if ackErr := redisq.Ack(ctx, constantPendingBlock, blockHeight); ackErr != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), ackErr.Error()))
				time.Sleep(2 * time.Second)
				return
			}
			finished = true
			delete(*retryBlockCount, blockHeightInt)
		} else {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", blockHeightInt, height)))

			retryCount := (*retryBlockCount)[blockHeightInt]
			retryCount++
			if retryCount >= setup.SweepThreshold {
				global.NODE_LOG.Error(fmt.Sprintf("give up block after retries: %d", blockHeightInt))
				if ackErr := redisq.Ack(ctx, constantPendingBlock, blockHeight); ackErr != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), ackErr.Error()))
					time.Sleep(2 * time.Second)
					return
				}
				finished = true
				delete(*retryBlockCount, blockHeightInt)
				return
			}
			(*retryBlockCount)[blockHeightInt] = retryCount
			time.Sleep(2 * time.Second)
			// defer 会 Requeue
		}
	}
}

func SweepBlockchainTransactionCoreForEthereum(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepBlockHeight int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) error {
	defer utils.HandlePanic()

	var err error
	var rpcBlockDetail response.RPCBlockInnerDetail
	client.URL = constant.GetAlchemyRPCUrlByChainId(chainId)
	payload := map[string]any{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "debug_traceBlockByNumber",
		"params": []any{
			"0x" + strconv.FormatInt(sweepBlockHeight, 16),
			map[string]any{
				"tracer": "callTracer",
			},
		},
	}

	err = client.HTTPPost(ctx, payload, &rpcBlockDetail)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	if len(rpcBlockDetail.Result) == 0 {
		err = errors.New("can not get the transaction of block number: " + fmt.Sprint(sweepBlockHeight))
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return err
	}

	monitorSet := utils.BuildHexAddressSet(*publicKey)
	if len(rpcBlockDetail.Result) > 0 {
		for _, transaction := range rpcBlockDetail.Result {
			isMonitorTx := false
			txFrom := utils.HexToAddress(transaction.Result.From)
			txTo := utils.HexToAddress(transaction.Result.To)

			matchArray := make([]string, 0)

			if transaction.Result.Error != "" {
				continue
			}

			if transaction.Result.Type == "CALL" {
				if transaction.Result.Input == "0x" {
					matchArray = append(matchArray, txFrom, txTo)
				} else {
					if isSupportContract, _, _, _ := sweepUtils.GetContractInfo(chainId, txTo); isSupportContract {
						if arrays, err := erc20.GetAllAddressByTransaction(chainId, txFrom, transaction.Hash, transaction.Result.Input); err == nil {
							matchArray = append(matchArray, arrays...)
						}
					}
				}
			}

			if len(transaction.Result.Calls) > 0 {
				matchArray = append(matchArray, processCallsForScanBlock(chainId, transaction.Hash, transaction.Result.Calls)...)
			}

			matchArray = utils.RemoveDuplicatesForString(matchArray)

			if len(matchArray) == 0 {
				continue
			}

		isMonitorTx = utils.MatchHexAddressSet(monitorSet, matchArray...)

			if isMonitorTx {
				var rpcReceipt response.RPCReceiptTransactionDetail
				var jsonRpcRequest request.JsonRpcRequest
				jsonRpcRequest.Id = 1
				jsonRpcRequest.Jsonrpc = "2.0"
				jsonRpcRequest.Method = "eth_getTransactionReceipt"
				jsonRpcRequest.Params = []any{transaction.Hash}

				err = client.HTTPPost(ctx, jsonRpcRequest, &rpcReceipt)
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					time.Sleep(2 * time.Second)
					return err
				}

				if rpcReceipt.Result.Status == "0x0" {
					continue
				}

				if _, err = redisq.EnqueueUnique(ctx, constantPendingTransaction, transaction.Hash); err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					time.Sleep(2 * time.Second)
					return err
				}
			}
		}
	}

	return nil
}

func handleInternalTransaction(ctx context.Context, client NODE_Client.Client, chainId uint, publicKey *[]string, notifyRequest request.NotificationRequest, rpcDetail response.RPCTransactionDetail) error {
	var err error

	var infos response.RPCInnerTxInfo
	client.URL = constant.GetAlchemyRPCUrlByChainId(chainId)
	payload := map[string]any{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "debug_traceTransaction",
		"params": []any{
			rpcDetail.Result.Hash,
			map[string]any{
				"tracer": "callTracer",
			},
		},
	}

	err = client.HTTPPost(ctx, payload, &infos)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s, hash: %s", constant.GetChainName(chainId), err.Error(), rpcDetail.Result.Hash))
		time.Sleep(2 * time.Second)
		return err
	}

	err = handleOrdinaryTransaction(ctx, chainId, publicKey, notifyRequest, infos.Result.From, infos.Result.To, rpcDetail.Result.Hash, infos.Result.Input, infos.Result.Value)
	topLevelOk := err == nil

	var callsOk bool
	if len(infos.Result.Calls) > 0 {
		callsOk = processCallsForSettle(ctx, chainId, publicKey, notifyRequest, rpcDetail.Result.Hash, infos.Result.Calls)
	}

	if !topLevelOk && !callsOk {
		return errors.New("not support")
	}

	return nil
}

func processCallsForSettle(ctx context.Context, chainId uint, publicKey *[]string, notifyRequest request.NotificationRequest, hash string, calls []response.CallResult) bool {
	isProcess := false

	var stack []response.CallResult
	stack = append(stack, calls...)

	for len(stack) > 0 {
		currentCall := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if currentCall.Type == "CALL" {
			if err := handleOrdinaryTransaction(ctx, chainId, publicKey, notifyRequest, currentCall.From, currentCall.To, hash, currentCall.Input, currentCall.Value); err == nil {
				isProcess = true
			}
		}

		if len(currentCall.Calls) > 0 {
			stack = append(stack, currentCall.Calls...)
		}
	}

	return isProcess
}

func processCallsForScanBlock(chainId uint, hash string, calls []response.CallResult) []string {
	matchArray := make([]string, 0)
	contractInfoCache := make(map[string]bool) // key: 地址, value: 是否支持

	var stack []response.CallResult
	stack = append(stack, calls...)

	for len(stack) > 0 {
		currentCall := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if currentCall.Error != "" {
			continue
		}

		if currentCall.Type == "CALL" {
			if currentCall.Input == "0x" {
				matchArray = append(matchArray, currentCall.From, currentCall.To)
			} else {
				addr := utils.HexToAddress(currentCall.To)

				isSupportContract, ok := contractInfoCache[addr]
				if !ok {
					isSupportContract, _, _, _ = sweepUtils.GetContractInfo(chainId, currentCall.To)
					contractInfoCache[addr] = isSupportContract
				}

				if isSupportContract {
					if arrays, err := erc20.GetAllAddressByTransaction(chainId, currentCall.From, hash, currentCall.Input); err == nil {
						matchArray = append(matchArray, arrays...)
					}
				}
			}
		}

		if len(currentCall.Calls) > 0 {
			stack = append(stack, currentCall.Calls...)
		}
	}
	return matchArray
}
