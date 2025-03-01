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
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func SetupTronLatestBlockHeight(client NODE_Client.Client, chainId uint) {
	var err error
	client.URL = constant.TronGetBlockByNetwork(chainId)
	client.Headers = map[string]string{
		"TRON-PRO-API-KEY": constant.GetRandomHTTPKeyByNetwork(chainId),
	}

	var blockRequest request.TronGetBlockRequest
	blockRequest.Detail = false
	var blockResponse response.TronGetBlockResponse
	err = client.HTTPPost(blockRequest, &blockResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	setup.SetupLatestBlockHeight(context.Background(), chainId, int64(blockResponse.BlockHeader.RawData.Number))
}

func SweepTronBlockchainTransaction(
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepCount *map[int64]int,
	sweepBlockHeight, cacheBlockHeight *int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) {
	defer utils.HandlePanic()

	if len(*publicKey) <= 0 {
		SetupTronLatestBlockHeight(client, chainId)
		setup.UpdateCacheBlockHeight(context.Background(), chainId)
		setup.UpdateSweepBlockHeight(context.Background(), chainId)
		setup.UpdatePublicKey(context.Background(), chainId)
		return
	}

	if *sweepBlockHeight > *cacheBlockHeight {
		SetupTronLatestBlockHeight(client, chainId)
		setup.UpdateCacheBlockHeight(context.Background(), chainId)
		setup.UpdatePublicKey(context.Background(), chainId)
		time.Sleep(time.Second * 3)
		return
	}

	var err error

	blockN, ok := (*sweepCount)[*sweepBlockHeight]
	if !ok {
		(*sweepCount)[*sweepBlockHeight] = 1
	} else if blockN >= setup.SweepThreshold {
		// skip current block
		_, err = global.NODE_REDIS.Set(context.Background(), constantSweepBlock, *sweepBlockHeight+1, 0).Result()
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return
		}

		// current block to pending queue
		_, err = global.NODE_REDIS.RPush(context.Background(), constantPendingBlock, *sweepBlockHeight).Result()
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return
		}

		delete(*sweepCount, *sweepBlockHeight)

		*sweepBlockHeight += 1
		return
	} else {
		(*sweepCount)[*sweepBlockHeight]++
	}

	client.URL = constant.TronGetBlockByNumByNetwork(chainId)
	client.Headers = map[string]string{
		"TRON-PRO-API-KEY": constant.GetRandomHTTPKeyByNetwork(chainId),
	}

	var blockByNumRequest request.TronGetBlockByNumRequest
	blockByNumRequest.Num = int(*sweepBlockHeight)
	var blockByNumResponse response.TronGetBlockByNumResponse
	err = client.HTTPPost(blockByNumRequest, &blockByNumResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	if *sweepBlockHeight == int64(blockByNumResponse.BlockHeader.RawData.Number) {

		if len(blockByNumResponse.Transactions) > 0 {
			for _, transaction := range blockByNumResponse.Transactions {

				if transaction.Ret[0].ContractRet != "SUCCESS" {
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
						// Determine duplicate transactions
						redisTxs, err := global.NODE_REDIS.LRange(context.Background(), constantPendingTransaction, 0, -1).Result()
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							return
						}

						for _, redisTx := range redisTxs {
							if redisTx == transaction.TxID {
								continue outerCurrentTxLoop
							}
						}

						_, err = global.NODE_REDIS.RPush(context.Background(), constantPendingTransaction, transaction.TxID).Result()
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							return
						}
						break
					}
				}

			}
		}

		_, err = global.NODE_REDIS.Set(context.Background(), constantSweepBlock, *sweepBlockHeight+1, 0).Result()
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return
		}

		*sweepBlockHeight += 1
	} else {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", *sweepBlockHeight, int64(blockByNumResponse.BlockHeader.RawData.Number))))
	}
}

func SweepTronBlockchainTransactionDetails(
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

	client.URL = constant.TronGetTxByIdByNetwork(chainId)
	client.Headers = map[string]string{
		"TRON-PRO-API-KEY": constant.GetRandomHTTPKeyByNetwork(chainId),
	}

	var txRequest request.TronGetBlockTxByIdRequest
	txRequest.Value = txHash
	var txResponse response.TronGetTxResponse
	err = client.HTTPPost(txRequest, &txResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	if txResponse.Ret[0].ContractRet != "SUCCESS" {
		return
	}

	var notifyRequest request.NotificationRequest
	notifyRequest.Hash = txResponse.TxID
	notifyRequest.Chain = chainId
	notifyRequest.BlockTimestamp = txResponse.RawData.Timestamp

	contractType := txResponse.RawData.Contract[0].Type

	switch contractType {
	case tron.TransferContract:
		err = handleTransferContractTx(chainId, publicKey, notifyRequest, txResponse)
	case tron.TriggerSmartContract:
		err = handleTriggerSmartContract(chainId, publicKey, notifyRequest, txResponse)
	default:
		return
	}

	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("Can not handle the tx: %s, Retry | %s -> %s", txHash, constant.GetChainName(chainId), err.Error()))
		return
	}

	_, err = global.NODE_REDIS.LPop(context.Background(), constantPendingTransaction).Result()
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}
}

func handleTransferContractTx(chainId uint, publicKey *[]string, notifyRequest request.NotificationRequest, txResponse response.TronGetTxResponse) error {
	var err error

	isSupportContract, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, "T9yD14Nj9j7xAB4dbGeiX9h8unkKHxuWwb")
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

	return handleNotification(chainId, publicKey, notifyRequest, fromAddress, toAddress)
}

func handleTriggerSmartContract(chainId uint, publicKey *[]string, notifyRequest request.NotificationRequest, txResponse response.TronGetTxResponse) error {
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

		return handleNotification(chainId, publicKey, notifyRequest, fromAddress, toAddress)

	case tron.TransferFrom:
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

		return handleNotification(chainId, publicKey, notifyRequest, fromAddress, toAddress)
	}

	return nil
}

func handleNotification(chainId uint, publicKey *[]string, notifyRequest request.NotificationRequest, fromAddress, toAddress string) error {
	var err error

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

			err = notification.NotificationRequest(notifyRequest)
			if err != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				return err
			}
		}

		if strings.EqualFold(v, toAddress) {
			notifyRequest.TransactType = "receive"
			notifyRequest.Address = v

			err = notification.NotificationRequest(notifyRequest)
			if err != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				return err
			}
		}
	}

	return nil
}

func SweepTronBlockchainPendingBlock(
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

	client.URL = constant.TronGetBlockByNumByNetwork(chainId)
	client.Headers = map[string]string{
		"TRON-PRO-API-KEY": constant.GetRandomHTTPKeyByNetwork(chainId),
	}

	var blockByNumRequest request.TronGetBlockByNumRequest
	blockByNumRequest.Num = int(blockHeightInt)
	var blockByNumResponse response.TronGetBlockByNumResponse
	err = client.HTTPPost(blockByNumRequest, &blockByNumResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	if blockHeightInt == int64(blockByNumResponse.BlockHeader.RawData.Number) {

		if len(blockByNumResponse.Transactions) > 0 {
			for _, transaction := range blockByNumResponse.Transactions {

				if transaction.Ret[0].ContractRet != "SUCCESS" {
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
						// Determine duplicate transactions
						redisTxs, err := global.NODE_REDIS.LRange(context.Background(), constantPendingTransaction, 0, -1).Result()
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							return
						}

						for _, redisTx := range redisTxs {
							if redisTx == transaction.TxID {
								continue outerCurrentTxLoop
							}
						}

						_, err = global.NODE_REDIS.RPush(context.Background(), constantPendingTransaction, transaction.TxID).Result()
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							return
						}
						break
					}
				}
			}
		}

		_, err = global.NODE_REDIS.LPop(context.Background(), constantPendingBlock).Result()
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		}
	} else {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", blockHeightInt, int64(blockByNumResponse.BlockHeader.RawData.Number))))
	}
}
