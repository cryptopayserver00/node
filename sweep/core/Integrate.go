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
	"strconv"
	"sync"
	"time"

	sweepUtils "node/sweep/utils"
	NODE_Client "node/utils/http"
	"node/utils/notification"

	"github.com/redis/go-redis/v9"
)

func SetupLatestBlockHeight(client NODE_Client.Client, chainId uint) {
	var err error
	client.URL = constant.GetRPCUrlByNetwork(chainId)
	var rpcBlockInfo response.RPCBlockInfo
	var jsonRpcRequest request.JsonRpcRequest
	jsonRpcRequest.Id = 1
	jsonRpcRequest.Jsonrpc = "2.0"
	jsonRpcRequest.Method = "eth_getBlockByNumber"
	jsonRpcRequest.Params = []interface{}{"latest", false}

	err = client.HTTPPost(jsonRpcRequest, &rpcBlockInfo)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
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
		setup.SetupLatestBlockHeight(context.Background(), chainId, int64(height))
	}
}

func SweepBlockchainTransaction(
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepCount *map[int64]int,
	sweepBlockHeight, cacheBlockHeight *int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) {
	defer utils.HandlePanic()

	if len(*publicKey) <= 0 {
		SetupLatestBlockHeight(client, chainId)
		setup.UpdateCacheBlockHeight(context.Background(), chainId)
		setup.UpdateSweepBlockHeight(context.Background(), chainId)
		setup.UpdatePublicKey(context.Background(), chainId)
		return
	}

	if *sweepBlockHeight >= *cacheBlockHeight {
		SetupLatestBlockHeight(client, chainId)
		setup.UpdateCacheBlockHeight(context.Background(), chainId)
		setup.UpdatePublicKey(context.Background(), chainId)
		time.Sleep(time.Second * 5)
		return
	}

	var wg sync.WaitGroup
	mutex := sync.Mutex{}

	var (
		numWorkers = 10
	)

	if chainId == constant.BSC_TESTNET {
		numWorkers = 3
	}

	if *sweepBlockHeight <= *cacheBlockHeight {
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

				if chainId == constant.ETH_MAINNET || chainId == constant.BSC_MAINNET || chainId == constant.ARBITRUM_ONE {
					err := SweepBlockchainTransactionCoreForEthereum(client, chainId, publicKey, sweepCount, currentHeight, constantSweepBlock, constantPendingBlock, constantPendingTransaction)
					if err != nil {

						_, err = global.NODE_REDIS.RPush(context.Background(), constantPendingBlock, currentHeight).Result()
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
						}
					}
				} else {
					err := SweepBlockchainTransactionCore(client, chainId, publicKey, sweepCount, currentHeight, constantSweepBlock, constantPendingBlock, constantPendingTransaction)
					if err != nil {

						_, err = global.NODE_REDIS.RPush(context.Background(), constantPendingBlock, currentHeight).Result()
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
						}
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

func SweepBlockchainTransactionCore(client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepCount *map[int64]int,
	sweepBlockHeight int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) error {
	defer utils.HandlePanic()

	var err error

	client.URL = constant.GetRPCUrlByNetwork(chainId)
	var rpcBlockDetail response.RPCBlockDetail
	var jsonRpcRequest request.JsonRpcRequest
	jsonRpcRequest.Id = 1
	jsonRpcRequest.Jsonrpc = "2.0"
	jsonRpcRequest.Method = "eth_getBlockByNumber"
	jsonRpcRequest.Params = []interface{}{"0x" + strconv.FormatInt(sweepBlockHeight, 16), true}

	err = client.HTTPPost(jsonRpcRequest, &rpcBlockDetail)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return err
	}

	if rpcBlockDetail.Result.Number == "" {
		err = fmt.Errorf("can not get the number: %s", rpcBlockDetail.Result.Number)
		return err
	}

	height, err := utils.HexStringToUint64(rpcBlockDetail.Result.Number)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return err
	}

	if sweepBlockHeight == int64(height) {
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
						if arrays, err := erc20.GetAllAddressByTransactionTwo(chainId, contractName, txFrom, txTo, transaction.Hash, transaction.Input); err == nil {
							matchArray = append(matchArray, arrays...)
						}
					}
				}

				matchArray = utils.RemoveDuplicatesForString(matchArray)

				if len(matchArray) == 0 {
					continue
				}

			outerCurrentTxLoop:
				for i := 0; i < len(*publicKey); i++ {
					for _, j := range matchArray {
						if utils.HexToAddress((*publicKey)[i]) == utils.HexToAddress(j) {
							isMonitorTx = true
							continue outerCurrentTxLoop
						}
					}
				}

				if isMonitorTx {
					// status of receipt
					var rpcReceipt response.RPCReceiptTransactionDetail
					jsonRpcRequest.Id = 1
					jsonRpcRequest.Jsonrpc = "2.0"
					jsonRpcRequest.Method = "eth_getTransactionReceipt"
					jsonRpcRequest.Params = []interface{}{transaction.Hash}

					err = client.HTTPPost(jsonRpcRequest, &rpcReceipt)
					if err != nil {
						global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
						return err
					}

					if rpcReceipt.Result.Status == "0x0" {
						continue
					}

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
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", sweepBlockHeight, height)))
		return errors.New("not the same height of block")
	}
}

func SweepBlockchainTransactionDetails(
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
	var rpcDetail response.RPCTransactionDetail
	var jsonRpcRequest request.JsonRpcRequest
	jsonRpcRequest.Id = 1
	jsonRpcRequest.Jsonrpc = "2.0"
	jsonRpcRequest.Method = "eth_getTransactionByHash"
	jsonRpcRequest.Params = []interface{}{txHash}

	err = client.HTTPPost(jsonRpcRequest, &rpcDetail)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	var rpcBlockInfo response.RPCBlockInfo
	jsonRpcRequest.Id = 1
	jsonRpcRequest.Jsonrpc = "2.0"
	jsonRpcRequest.Method = "eth_getBlockByNumber"
	jsonRpcRequest.Params = []interface{}{rpcDetail.Result.BlockNumber, false}

	err = client.HTTPPost(jsonRpcRequest, &rpcBlockInfo)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	// handle drop tx
	if rpcBlockInfo.Result.Timestamp == "" {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), "Can not handle the drop tx: "+txHash))

		_, err = global.NODE_REDIS.LPop(context.Background(), constantPendingTransaction).Result()
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		}
		return
	}

	blockTimeStamp, err := utils.HexStringToUint64(rpcBlockInfo.Result.Timestamp)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	var notifyRequest request.NotificationRequest

	notifyRequest.Hash = rpcDetail.Result.Hash
	notifyRequest.Chain = chainId
	notifyRequest.BlockTimestamp = int(blockTimeStamp) * 1000

	var isProcess = false

	if chainId == constant.ETH_MAINNET || chainId == constant.BSC_MAINNET || chainId == constant.ARBITRUM_ONE {
		isProcess, err = handleEthereumTx(client, chainId, publicKey, notifyRequest, rpcDetail)
	} else {
		if rpcDetail.Result.Input == "0x" {
			isProcess, err = handleERC20(chainId, publicKey, notifyRequest, rpcDetail.Result.From, rpcDetail.Result.To, rpcDetail.Result.Hash, rpcDetail.Result.Input, rpcDetail.Result.Value)
		} else {
			_, contractName, _, _ := sweepUtils.GetContractInfo(chainId, rpcDetail.Result.To)

			switch contractName {
			default:
				isProcess, err = handleERC20(chainId, publicKey, notifyRequest, rpcDetail.Result.From, rpcDetail.Result.To, rpcDetail.Result.Hash, rpcDetail.Result.Input, rpcDetail.Result.Value)
			}
		}
	}

	if !isProcess || err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("Can not handle the tx: %s, Retry | %s -> %s", rpcDetail.Result.Hash, constant.GetChainName(chainId), err.Error()))
		return
	}

	_, err = global.NODE_REDIS.LPop(context.Background(), constantPendingTransaction).Result()
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}
}

func handleERC20(chainId uint, publicKey *[]string, notifyRequest request.NotificationRequest, from, to, hash, input, value string) (isProcess bool, err error) {

	var (
		isSupportContract bool
		contractName      string
		decimals          int
	)

	if input == "0x" {
		isSupportContract, contractName, _, decimals = sweepUtils.GetContractInfo(chainId, "0x0000000000000000000000000000000000000000")
	} else {
		isSupportContract, contractName, _, decimals = sweepUtils.GetContractInfo(chainId, to)
	}

	if !isSupportContract {
		err = errors.New("can not find the contract: " + to)
		isProcess = false
		return
	}

	fromAddress := from
	notifyRequest.FromAddress = fromAddress

	if decimals == 0 {
		err = errors.New("decimals can not be 0")
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		isProcess = false
		return
	}

	if !(input == "0x") {
		methodName, decodeFromAddress, decodeToAddress, transactionValue, err := erc20.DecodeERC20TransactionInputData(chainId, hash, input)
		if err != nil {
			isProcess = false
			return isProcess, err
		}

		switch methodName {
		case erc20.TransferFrom:
			fromAddress = decodeFromAddress
			notifyRequest.FromAddress = fromAddress
		}

		notifyRequest.ToAddress = decodeToAddress
		notifyRequest.Token = contractName
		notifyRequest.Amount = utils.CalculateBalance(transactionValue, decimals)

		for _, v := range *publicKey {
			if utils.HexToAddress(fromAddress) == utils.HexToAddress(v) {
				notifyRequest.TransactType = "send"
				notifyRequest.Address = v

				err = notification.NotificationRequest(notifyRequest)
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					isProcess = false
					return isProcess, err
				}
				isProcess = true
			}

			if utils.HexToAddress(decodeToAddress) == utils.HexToAddress(v) {
				notifyRequest.TransactType = "receive"
				notifyRequest.Address = v

				err = notification.NotificationRequest(notifyRequest)
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					isProcess = false
					return isProcess, err
				}
				isProcess = true
			}
		}

	} else {
		toAddress := to
		notifyRequest.ToAddress = toAddress

		value, err := utils.HexStringToBigInt(value)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			isProcess = false
			return isProcess, err
		}
		notifyRequest.Amount = utils.CalculateBalance(value, decimals)
		notifyRequest.Token = contractName

		for _, v := range *publicKey {
			if utils.HexToAddress(fromAddress) == utils.HexToAddress(v) {

				notifyRequest.TransactType = "send"
				notifyRequest.Address = v

				err = notification.NotificationRequest(notifyRequest)
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					isProcess = false
					return isProcess, err
				}
				isProcess = true
			}

			if utils.HexToAddress(toAddress) == utils.HexToAddress(v) {
				notifyRequest.TransactType = "receive"
				notifyRequest.Address = v

				err = notification.NotificationRequest(notifyRequest)
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					isProcess = false
					return isProcess, err
				}
				isProcess = true
			}
		}
	}

	return isProcess, nil
}

func SweepBlockchainPendingBlock(
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepCount *map[int64]int,
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

	if chainId == constant.ETH_MAINNET || chainId == constant.BSC_MAINNET || chainId == constant.ARBITRUM_ONE {
		var rpcBlockDetail response.RPCBlockInnerDetail
		client.URL = constant.GetInnerTxRPCUrlByNetwork(chainId)
		payload := map[string]interface{}{
			"id":      1,
			"jsonrpc": "2.0",
			"method":  "debug_traceBlockByNumber",
			"params": []interface{}{
				"0x" + strconv.FormatInt(blockHeightInt, 16),
				map[string]interface{}{
					"tracer": "callTracer",
				},
			},
		}

		err = client.HTTPPost(payload, &rpcBlockDetail)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return
		}

		if len(rpcBlockDetail.Result) == 0 {
			blockN, ok := (*sweepCount)[blockHeightInt]
			if !ok {
				(*sweepCount)[blockHeightInt] = 1

				err = errors.New("can not get the transaction of block number: " + fmt.Sprint(blockHeightInt))
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				return
			} else if blockN >= setup.SweepThreshold {
				delete(*sweepCount, blockHeightInt)
			} else {
				(*sweepCount)[blockHeightInt]++

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

			outerETHCurrentTxPool:
				for i := 0; i < len(*publicKey); i++ {
					for _, j := range matchArray {
						if utils.HexToAddress((*publicKey)[i]) == utils.HexToAddress(j) {
							isMonitorTx = true
							continue outerETHCurrentTxPool
						}
					}
				}

				if isMonitorTx {
					// status of receipt
					var rpcReceipt response.RPCReceiptTransactionDetail
					var jsonRpcRequest request.JsonRpcRequest
					jsonRpcRequest.Id = 1
					jsonRpcRequest.Jsonrpc = "2.0"
					jsonRpcRequest.Method = "eth_getTransactionReceipt"
					jsonRpcRequest.Params = []interface{}{transaction.Hash}

					err = client.HTTPPost(jsonRpcRequest, &rpcReceipt)
					if err != nil {
						global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
						return
					}

					if rpcReceipt.Result.Status == "0x0" {
						continue
					}

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

		*sweepCount = make(map[int64]int)
	} else {
		client.URL = constant.GetRPCUrlByNetwork(chainId)
		var rpcBlockDetail response.RPCBlockDetail
		var jsonRpcRequest request.JsonRpcRequest
		jsonRpcRequest.Id = 1
		jsonRpcRequest.Jsonrpc = "2.0"
		jsonRpcRequest.Method = "eth_getBlockByNumber"
		jsonRpcRequest.Params = []interface{}{"0x" + strconv.FormatInt(blockHeightInt, 16), true}

		err = client.HTTPPost(jsonRpcRequest, &rpcBlockDetail)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return
		}

		if rpcBlockDetail.Result.Number == "" {
			err = fmt.Errorf("can not get the number: %s, %s", blockHeight, client.URL)
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return
		}

		height, err := utils.HexStringToUint64(rpcBlockDetail.Result.Number)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
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
							if arrays, err := erc20.GetAllAddressByTransactionTwo(chainId, contractName, txFrom, txTo, transaction.Hash, transaction.Input); err == nil {
								matchArray = append(matchArray, arrays...)
							}
						}
					}

					matchArray = utils.RemoveDuplicatesForString(matchArray)

					if len(matchArray) == 0 {
						continue
					}

				outerCurrentTxLoop:
					for i := 0; i < len(*publicKey); i++ {
						for _, j := range matchArray {
							if utils.HexToAddress((*publicKey)[i]) == utils.HexToAddress(j) {
								isMonitorTx = true
								continue outerCurrentTxLoop
							}
						}
					}

					if isMonitorTx {
						// status of receipt
						var rpcReceipt response.RPCReceiptTransactionDetail
						jsonRpcRequest.Id = 1
						jsonRpcRequest.Jsonrpc = "2.0"
						jsonRpcRequest.Method = "eth_getTransactionReceipt"
						jsonRpcRequest.Params = []interface{}{transaction.Hash}

						err = client.HTTPPost(jsonRpcRequest, &rpcReceipt)
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							return
						}

						if rpcReceipt.Result.Status == "0x0" {
							continue
						}

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
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", blockHeightInt, height)))
		}
	}
}

func SweepBlockchainTransactionCoreForEthereum(client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepCount *map[int64]int,
	sweepBlockHeight int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) error {
	defer utils.HandlePanic()

	var err error
	var rpcBlockDetail response.RPCBlockInnerDetail
	client.URL = constant.GetInnerTxRPCUrlByNetwork(chainId)
	payload := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "debug_traceBlockByNumber",
		"params": []interface{}{
			"0x" + strconv.FormatInt(sweepBlockHeight, 16),
			map[string]interface{}{
				"tracer": "callTracer",
			},
		},
	}

	err = client.HTTPPost(payload, &rpcBlockDetail)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return err
	}

	if len(rpcBlockDetail.Result) == 0 {
		err = errors.New("can not get the transaction of block number: " + fmt.Sprint(sweepBlockHeight))
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return err
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

		outerCurrentTxLoop:
			for i := 0; i < len(*publicKey); i++ {
				for _, j := range matchArray {
					if utils.HexToAddress((*publicKey)[i]) == utils.HexToAddress(j) {
						isMonitorTx = true
						continue outerCurrentTxLoop
					}
				}
			}

			if isMonitorTx {
				// status of receipt
				var rpcReceipt response.RPCReceiptTransactionDetail
				var jsonRpcRequest request.JsonRpcRequest
				jsonRpcRequest.Id = 1
				jsonRpcRequest.Jsonrpc = "2.0"
				jsonRpcRequest.Method = "eth_getTransactionReceipt"
				jsonRpcRequest.Params = []interface{}{transaction.Hash}

				err = client.HTTPPost(jsonRpcRequest, &rpcReceipt)
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					return err
				}

				if rpcReceipt.Result.Status == "0x0" {
					continue
				}

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
}

func handleEthereumTx(client NODE_Client.Client, chainId uint, publicKey *[]string, notifyRequest request.NotificationRequest, rpcDetail response.RPCTransactionDetail) (isProcess bool, err error) {

	var infos response.RPCInnerTxInfo
	client.URL = constant.GetInnerTxRPCUrlByNetwork(chainId)
	payload := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "debug_traceTransaction",
		"params": []interface{}{
			rpcDetail.Result.Hash,
			map[string]interface{}{
				"tracer": "callTracer",
			},
		},
	}

	err = client.HTTPPost(payload, &infos)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s, hash: %s", constant.GetChainName(chainId), err.Error(), rpcDetail.Result.Hash))
		isProcess = false
		return
	}

	isProcess, _ = handleERC20(chainId, publicKey, notifyRequest, infos.Result.From, infos.Result.To, rpcDetail.Result.Hash, infos.Result.Input, infos.Result.Value)

	if len(infos.Result.Calls) > 0 {
		if processCallsForSettle(chainId, publicKey, notifyRequest, rpcDetail.Result.Hash, infos.Result.Calls) {
			isProcess = true
		}
	}

	if !isProcess {
		err = fmt.Errorf("can not handle the tx: %s", rpcDetail.Result.Hash)
		return
	}

	global.NODE_LOG.Info(fmt.Sprintf("successful handling, tx: %s", rpcDetail.Result.Hash))

	return
}

func processCallsForSettle(chainId uint, publicKey *[]string, notifyRequest request.NotificationRequest, hash string, calls []response.CallResult) bool {
	isProcess := false

	var stack []response.CallResult
	stack = append(stack, calls...)

	for len(stack) > 0 {
		currentCall := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if currentCall.Type == "CALL" {
			process, _ := handleERC20(chainId, publicKey, notifyRequest, currentCall.From, currentCall.To, hash, currentCall.Input, currentCall.Value)
			if process {
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
	contractInfoCache := make(map[string]bool)

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
				if isSupport := contractInfoCache[utils.HexToAddress(currentCall.To)]; isSupport {
					if arrays, err := erc20.GetAllAddressByTransaction(chainId, currentCall.From, hash, currentCall.Input); err == nil {
						matchArray = append(matchArray, arrays...)
					}
				} else {
					if isSupportContract, _, _, _ := sweepUtils.GetContractInfo(chainId, currentCall.To); isSupportContract {
						if arrays, err := erc20.GetAllAddressByTransaction(chainId, currentCall.From, hash, currentCall.Input); err == nil {
							matchArray = append(matchArray, arrays...)
						}

						contractInfoCache[utils.HexToAddress(currentCall.To)] = true
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
