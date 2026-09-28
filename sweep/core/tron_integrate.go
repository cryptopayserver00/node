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
	sweepUtils "node/sweep/utils"
	"node/sweep/utils/tron"
	"node/utils"
	NODE_Client "node/utils/http"
	"node/utils/notification"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

func SetupTronLatestBlockHeight(ctx context.Context, client NODE_Client.Client, chainId uint) {
	var err error
	client.URL = constant.TronGetBlockByNetwork(chainId)
	client.Headers = map[string]string{
		"TRON-PRO-API-KEY": constant.GetRandomHTTPKeyByNetwork(chainId),
	}

	var blockRequest request.TronGetBlockRequest
	blockRequest.Detail = false
	var blockResponse response.TronGetBlockResponse
	err = client.HTTPPost(ctx, blockRequest, &blockResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return
	}

	if blockResponse.BlockHeader.RawData.Number > 0 {
		setup.SetupLatestBlockHeight(ctx, chainId, int64(blockResponse.BlockHeader.RawData.Number))
	}
}

func SweepTronBlockchainTransaction(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepBlockHeight, cacheBlockHeight *int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) {
	defer utils.HandlePanic()

	if len(*publicKey) <= 0 {
		SetupTronLatestBlockHeight(ctx, client, chainId)
		setup.UpdateCacheBlockHeight(ctx, chainId)
		setup.UpdateSweepBlockHeight(ctx, chainId)
		setup.UpdatePublicKey(ctx, chainId)
		return
	}

	if *sweepBlockHeight >= *cacheBlockHeight {
		SetupTronLatestBlockHeight(ctx, client, chainId)
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

			if err := SweepTronBlockchainTransactionCore(ctx, client, chainId, publicKey, height, constantSweepBlock, constantPendingBlock, constantPendingTransaction); err != nil {
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

func SweepTronBlockchainTransactionCore(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepBlockHeight int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) error {
	defer utils.HandlePanic()

	var err error

	client.URL = constant.TronGetBlockByNumByNetwork(chainId)
	client.Headers = map[string]string{
		"TRON-PRO-API-KEY": constant.GetRandomHTTPKeyByNetwork(chainId),
	}

	var blockByNumRequest request.TronGetBlockByNumRequest
	blockByNumRequest.Num = int(sweepBlockHeight)
	var blockByNumResponse response.TronGetBlockByNumResponse
	err = client.HTTPPost(ctx, blockByNumRequest, &blockByNumResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	if sweepBlockHeight == int64(blockByNumResponse.BlockHeader.RawData.Number) {
		if len(blockByNumResponse.Transactions) > 0 {
			for _, transaction := range blockByNumResponse.Transactions {
				if len(transaction.Ret) == 0 {
					continue
				}

				if transaction.Ret[0].ContractRet != "SUCCESS" {
					continue
				}

				if len(transaction.RawData.Contract) == 0 {
					continue
				}

				contractType := transaction.RawData.Contract[0].Type
				var toAddress string

				if contractType == tron.TransferContract {
					toAddress = transaction.RawData.Contract[0].Parameter.Value.ToAddress
				} else if contractType == tron.TriggerSmartContract {
					contractData := transaction.RawData.Contract[0].Parameter.Value.Data
					methodID, _, _ := tron.TronDecodeMethod(contractData)

					method := tron.KnownMethods[methodID]
					if method == "" {
						continue
					}

					toAddress = transaction.RawData.Contract[0].Parameter.Value.ContractAddress

					adds, err := tron.FromHexAddress(toAddress)
					if err != nil {
						continue
					}

					if isSupportContract, _, _, _ := sweepUtils.GetContractInfo(chainId, adds); !isSupportContract {
						continue
					}

				} else {
					continue
				}

				isMonitorTx := false

			outerCurrentTxLoop:

				for i := 0; i < len(*publicKey); i++ {
					isMonitorTx = tron.IsHandleTransaction(
						chainId,
						transaction.TxID,
						contractType,
						transaction.RawData.Contract[0].Parameter.Value.OwnerAddress,
						toAddress,
						(*publicKey)[i],
						transaction.RawData.Contract[0].Parameter.Value.Data,
					)

					if isMonitorTx {
						redisTxs, err := global.NODE_REDIS.LRange(ctx, constantPendingTransaction, 0, -1).Result()
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							time.Sleep(2 * time.Second)
							return err
						}

						found := slices.Contains(redisTxs, transaction.TxID)
						if found {
							break outerCurrentTxLoop
						} else {
							_, err = global.NODE_REDIS.RPush(ctx, constantPendingTransaction, transaction.TxID).Result()
							if err != nil {
								global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
								time.Sleep(2 * time.Second)
								return err
							}
							break
						}
					}
				}
			}
		}

		return nil
	} else {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", sweepBlockHeight, int64(blockByNumResponse.BlockHeader.RawData.Number))))
		time.Sleep(2 * time.Second)
		return errors.New("not the same height of block")
	}
}

func SweepTronBlockchainTransactionDetails(
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

	err = handleTronTransactionDetails(ctx, client, chainId, publicKey, txHash)

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

func handleTronTransactionDetails(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	txHash string,
) error {
	var err error

	client.URL = constant.TronGetTxByIdByNetwork(chainId)
	client.Headers = map[string]string{
		"TRON-PRO-API-KEY": constant.GetRandomHTTPKeyByNetwork(chainId),
	}

	var txRequest request.TronGetBlockTxByIdRequest
	txRequest.Value = txHash
	var txResponse response.TronGetTxResponse
	err = client.HTTPPost(ctx, txRequest, &txResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	if len(txResponse.Ret) == 0 {
		return errors.New("not support")
	}

	if txResponse.Ret[0].ContractRet != "SUCCESS" {
		return errors.New("not support")
	}

	if len(txResponse.RawData.Contract) == 0 {
		return errors.New("not support")
	}

	var notifyRequest request.NotificationRequest
	notifyRequest.Hash = txResponse.TxID
	notifyRequest.Chain = chainId
	notifyRequest.BlockTimestamp = txResponse.RawData.Timestamp

	contractType := txResponse.RawData.Contract[0].Type

	switch contractType {
	case tron.TransferContract:
		err = handleTransferContractTx(ctx, chainId, publicKey, notifyRequest, txResponse)
	case tron.TriggerSmartContract:
		err = handleTriggerSmartContract(ctx, chainId, publicKey, notifyRequest, txResponse)
	default:
		return errors.New("not support")
	}

	return err
}

func handleTransferContractTx(ctx context.Context, chainId uint, publicKey *[]string, notifyRequest request.NotificationRequest, txResponse response.TronGetTxResponse) error {
	var err error

	isSupportContract, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, constant.TRX_NATIVE_PLACEHOLDER_ADDRESS)
	if !isSupportContract {
		err = errors.New("can not find the contract")
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return err
	}
	if decimals == 0 {
		err = errors.New("decimals can not be 0")
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return err
	}

	fromAddress, err := tron.FromHexAddress(txResponse.RawData.Contract[0].Parameter.Value.OwnerAddress)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return err
	}
	toAddress, err := tron.FromHexAddress(txResponse.RawData.Contract[0].Parameter.Value.ToAddress)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return err
	}

	notifyRequest.Amount = utils.CalculateBalance(big.NewInt(int64(txResponse.RawData.Contract[0].Parameter.Value.Amount)), decimals)
	notifyRequest.Token = contractName

	return handleNotification(ctx, chainId, publicKey, notifyRequest, fromAddress, toAddress)
}

func handleTriggerSmartContract(ctx context.Context, chainId uint, publicKey *[]string, notifyRequest request.NotificationRequest, txResponse response.TronGetTxResponse) error {
	contractData := txResponse.RawData.Contract[0].Parameter.Value.Data
	methodID, _, _ := tron.TronDecodeMethod(contractData)

	method := tron.KnownMethods[methodID]
	if method == "" {
		err := fmt.Errorf("can not find the method: %s", methodID)
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return err
	}

	switch method {
	case tron.Transfer:
		fromAddress, err := tron.FromHexAddress(txResponse.RawData.Contract[0].Parameter.Value.OwnerAddress)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return err
		}

		const minTransferDataLen = 136 // 8(selector) + 64(to slot) + 64(value slot)
		if len(contractData) < minTransferDataLen {
			err := fmt.Errorf("insufficient contract data length: %d", len(contractData))
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return err
		}

		toAddress, err := tron.FromHexAddress("41" + contractData[32:72])
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return err
		}

		value, good := new(big.Int).SetString(contractData[len(contractData)-64:], 16)
		if !good {
			err = fmt.Errorf("can not decode the value: %s", contractData[len(contractData)-64:])
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return err
		}

		contractAddress, err := tron.FromHexAddress(txResponse.RawData.Contract[0].Parameter.Value.ContractAddress)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return err
		}

		_, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, contractAddress)

		notifyRequest.Token = contractName
		notifyRequest.Amount = utils.CalculateBalance(value, decimals)

		return handleNotification(ctx, chainId, publicKey, notifyRequest, fromAddress, toAddress)
	case tron.TransferFrom:
		const minTransferDataLen = 200 // 8(selector) + 64(from slot) + 64(to slot) + 64(value slot)
		if len(contractData) < minTransferDataLen {
			err := fmt.Errorf("insufficient contract data length: %d", len(contractData))
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return err
		}

		fromAddress, err := tron.FromHexAddress("41" + contractData[32:72])
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return err
		}

		toAddress, err := tron.FromHexAddress("41" + contractData[96:136])
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return err
		}

		value, good := new(big.Int).SetString(contractData[len(contractData)-64:], 16)
		if !good {
			err = fmt.Errorf("can not decode the value: %s", contractData[len(contractData)-64:])
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return err
		}

		contractAddress, err := tron.FromHexAddress(txResponse.RawData.Contract[0].Parameter.Value.ContractAddress)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return err
		}

		_, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, contractAddress)

		notifyRequest.Token = contractName
		notifyRequest.Amount = utils.CalculateBalance(value, decimals)

		return handleNotification(ctx, chainId, publicKey, notifyRequest, fromAddress, toAddress)
	default:
		err := fmt.Errorf("known but unhandled method: %s", method)
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return err
	}
}

func handleNotification(ctx context.Context, chainId uint, publicKey *[]string, notifyRequest request.NotificationRequest, fromAddress, toAddress string) error {
	var err error
	isProcess := false

	if fromAddress == "" || toAddress == "" {
		err = fmt.Errorf("can not be empty, fromAddress: %s, toAddress: %s", fromAddress, toAddress)
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return err
	}

	notifyRequest.FromAddress = fromAddress
	notifyRequest.ToAddress = toAddress

	for _, v := range *publicKey {
		if strings.EqualFold(v, fromAddress) {
			notifyRequest.TransactType = "send"
			notifyRequest.Address = v

			err = notification.NotificationRequest(ctx, notifyRequest)
			if err != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				return err
			}
			isProcess = true
		}

		if strings.EqualFold(v, toAddress) {
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

func SweepTronBlockchainPendingBlock(
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

	client.URL = constant.TronGetBlockByNumByNetwork(chainId)
	client.Headers = map[string]string{
		"TRON-PRO-API-KEY": constant.GetRandomHTTPKeyByNetwork(chainId),
	}

	var blockByNumRequest request.TronGetBlockByNumRequest
	blockByNumRequest.Num = int(blockHeightInt)
	var blockByNumResponse response.TronGetBlockByNumResponse
	err = client.HTTPPost(ctx, blockByNumRequest, &blockByNumResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return
	}

	if blockHeightInt == int64(blockByNumResponse.BlockHeader.RawData.Number) {
		if len(blockByNumResponse.Transactions) > 0 {
			for _, transaction := range blockByNumResponse.Transactions {
				if len(transaction.Ret) == 0 {
					continue
				}

				if transaction.Ret[0].ContractRet != "SUCCESS" {
					continue
				}

				if len(transaction.RawData.Contract) == 0 {
					continue
				}

				contractType := transaction.RawData.Contract[0].Type
				var toAddress string

				if contractType == tron.TransferContract {
					toAddress = transaction.RawData.Contract[0].Parameter.Value.ToAddress
				} else if contractType == tron.TriggerSmartContract {
					contractData := transaction.RawData.Contract[0].Parameter.Value.Data
					methodID, _, _ := tron.TronDecodeMethod(contractData)

					method := tron.KnownMethods[methodID]
					if method == "" {
						continue
					}

					toAddress = transaction.RawData.Contract[0].Parameter.Value.ContractAddress

					adds, err := tron.FromHexAddress(toAddress)
					if err != nil {
						continue
					}

					if isSupportContract, _, _, _ := sweepUtils.GetContractInfo(chainId, adds); !isSupportContract {
						continue
					}
				} else {
					continue
				}

				isMonitorTx := false

			outerCurrentTxLoop:

				for i := 0; i < len(*publicKey); i++ {
					isMonitorTx = tron.IsHandleTransaction(
						chainId,
						transaction.TxID,
						contractType,
						transaction.RawData.Contract[0].Parameter.Value.OwnerAddress,
						toAddress,
						(*publicKey)[i],
						transaction.RawData.Contract[0].Parameter.Value.Data,
					)

					if isMonitorTx {
						redisTxs, err := global.NODE_REDIS.LRange(ctx, constantPendingTransaction, 0, -1).Result()
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							return
						}

						found := slices.Contains(redisTxs, transaction.TxID)
						if found {
							break outerCurrentTxLoop
						} else {
							_, err = global.NODE_REDIS.RPush(ctx, constantPendingTransaction, transaction.TxID).Result()
							if err != nil {
								global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
								return
							}
							break
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
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", blockHeightInt, int64(blockByNumResponse.BlockHeader.RawData.Number))))

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
