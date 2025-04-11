package plugin

import (
	"context"
	"fmt"
	"math/big"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response/mempool"
	sweepUtils "node/sweep/utils"
	"node/sweep/utils/btc"
	"node/utils"
	NODE_Client "node/utils/http"
	"node/utils/notification"
	"strconv"
	"strings"
)

func GetBchBlockHeightByMempool(client NODE_Client.Client, chainId uint) int64 {
	var err error
	client.URL = constant.MempoolGetBlockHeightByNetwork(chainId)
	var blockHeight int64
	err = client.HTTPGetUnique(&blockHeight)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return 0
	}

	return blockHeight
}

func HandleBchBlockTransactionsByMempool(
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepCount *map[int64]int,
	sweepBlockHeight *int64,
	constantSweepBlock, constantPendingTransaction string,
) {
	var err error

	var blockHash string
	client.URL = fmt.Sprintf(constant.MempoolGetBlockHashByNetwork(chainId), *sweepBlockHeight)
	err = client.HTTPGetUnique(&blockHash)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	var block mempool.MempoolBlock
	client.URL = fmt.Sprintf(constant.MempoolGetBlockByNetwork(chainId), blockHash)
	err = client.HTTPGet(&block)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	var bitcoincashTxsResponses []mempool.MempoolTx

	for i := 0; i < block.TxCount; i += 25 {
		client.URL = fmt.Sprintf(constant.MempoolGetBlockTransactionByNetwork(chainId), blockHash, i)
		var bitcoincashTxsResponse []mempool.MempoolTx
		err = client.HTTPGet(&bitcoincashTxsResponse)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return
		}

		bitcoincashTxsResponses = append(bitcoincashTxsResponses, bitcoincashTxsResponse...)
	}

	if len(bitcoincashTxsResponses) == 0 {
		return
	}

	if *sweepBlockHeight == int64(block.Height) {

		if len(bitcoincashTxsResponses) > 0 {
			for _, transaction := range bitcoincashTxsResponses {

				if len(transaction.Vin) == 0 || len(transaction.Vout) == 0 {
					continue
				}

			outerCurrentTxLoop:
				for i := 0; i < len(*publicKey); i++ {
					isMonitorTx := false

					if len(transaction.Vin) > 0 {
						for _, input := range transaction.Vin {
							if strings.EqualFold((*publicKey)[i], input.Prevout.Scriptpubkey_address) {
								isMonitorTx = true
								break
							}
						}
					}

					if len(transaction.Vout) > 0 {
						for _, output := range transaction.Vout {
							if strings.EqualFold((*publicKey)[i], output.Scriptpubkey_address) {
								isMonitorTx = true
								break
							}
						}
					}

					if isMonitorTx {

						// Determine duplicate transactions
						redisTxs, err := global.NODE_REDIS.LRange(context.Background(), constantPendingTransaction, 0, -1).Result()
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							return
						}

						for _, redisTx := range redisTxs {
							if redisTx == transaction.TxId {
								continue outerCurrentTxLoop
							}
						}

						_, err = global.NODE_REDIS.RPush(context.Background(), constantPendingTransaction, transaction.TxId).Result()
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

		delete(*sweepCount, *sweepBlockHeight)

		*sweepBlockHeight += 1
	} else {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same sweepBlockHeight and blockHeight: %d - %d", *sweepBlockHeight, int64(block.Height))))
	}
}

func HandleBchTransactionDetailsByMempool(
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	constantPendingTransaction string,
	txHash string,
) {

	global.NODE_LOG.Info(fmt.Sprintf("%s -> handle mempool detail: %s", constant.GetChainName(chainId), txHash))

	var err error
	var isProcess bool

	client.URL = fmt.Sprintf(constant.MempoolGetTransctionByNetwork(chainId), txHash)

	var bitcoincashTxResponse mempool.MempoolTx
	err = client.HTTPGet(&bitcoincashTxResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	var notifyRequest request.NotificationRequest

	notifyRequest.Hash = bitcoincashTxResponse.TxId
	notifyRequest.Chain = chainId
	notifyRequest.BlockTimestamp = bitcoincashTxResponse.Status.BlockTime * 1000

	if len(bitcoincashTxResponse.Vin) == 0 || len(bitcoincashTxResponse.Vout) == 0 {
		return
	}

	_, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, "")
	if decimals == 0 {
		return
	}

	for _, input := range bitcoincashTxResponse.Vin {
		if input.Prevout.Scriptpubkey_address != "" {
			notifyRequest.FromAddress = input.Prevout.Scriptpubkey_address
			continue
		}
	}

	var isOnmiUSDT bool
	var onmiData map[string]int

	for _, output := range bitcoincashTxResponse.Vout {
		if output.Value == 0 && output.Scriptpubkey_address == "" {
			onmiData, isOnmiUSDT = btc.ParseOmniUSDTData(output.Scriptpubkey)
			break
		} else {
			isOnmiUSDT = false
		}
	}

	if isOnmiUSDT {
		// omni
		notifyRequest.Token = "USDT"
		notifyRequest.Amount = strconv.Itoa(onmiData["token_amount"])

		for _, omniOutput := range bitcoincashTxResponse.Vout {
			if strings.EqualFold(omniOutput.Scriptpubkey_address, notifyRequest.FromAddress) || omniOutput.Value == 0 || omniOutput.Scriptpubkey_address == "" {
				continue
			}
			notifyRequest.ToAddress = omniOutput.Scriptpubkey_address
		}

		for _, v := range *publicKey {
			notifyRequest.Address = v

			if strings.EqualFold(notifyRequest.FromAddress, v) {
				notifyRequest.TransactType = "send"

				err = notification.NotificationRequest(notifyRequest)
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					return
				}
				isProcess = true
			}

			if strings.EqualFold(notifyRequest.ToAddress, v) {
				notifyRequest.TransactType = "receive"

				err = notification.NotificationRequest(notifyRequest)
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					return
				}
				isProcess = true
			}
		}

	} else {
		notifyRequest.Token = contractName
		for _, output := range bitcoincashTxResponse.Vout {
			if strings.EqualFold(output.Scriptpubkey_address, notifyRequest.FromAddress) {
				continue
			}

			notifyRequest.Amount = utils.CalculateBalance(big.NewInt(int64(output.Value)), decimals)
			for _, v := range *publicKey {
				notifyRequest.Address = v
				notifyRequest.ToAddress = output.Scriptpubkey_address

				if strings.EqualFold(notifyRequest.FromAddress, v) {
					notifyRequest.TransactType = "send"

					err = notification.NotificationRequest(notifyRequest)
					if err != nil {
						global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
						return
					}
					isProcess = true
				}

				if strings.EqualFold(output.Scriptpubkey_address, v) {
					notifyRequest.TransactType = "receive"

					err = notification.NotificationRequest(notifyRequest)
					if err != nil {
						global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
						return
					}
					isProcess = true
				}
			}
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

func HandleBchPendingBlockByMempool(
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	constantPendingBlock, constantPendingTransaction string,
	blockHeight string,
	blockHeightInt int64,
) {
	var err error

	var blockHash string
	client.URL = fmt.Sprintf(constant.MempoolGetBlockHashByNetwork(chainId), blockHeightInt)
	err = client.HTTPGetUnique(&blockHash)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	var block mempool.MempoolBlock
	client.URL = fmt.Sprintf(constant.MempoolGetBlockByNetwork(chainId), blockHash)
	err = client.HTTPGet(&block)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	var bitcoincashTxsResponses []mempool.MempoolTx

	for i := 0; i < block.TxCount; i += 25 {
		client.URL = fmt.Sprintf(constant.MempoolGetBlockTransactionByNetwork(chainId), blockHash, i)
		var bitcoinTxsResponse []mempool.MempoolTx
		err = client.HTTPGet(&bitcoinTxsResponse)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return
		}

		bitcoincashTxsResponses = append(bitcoincashTxsResponses, bitcoinTxsResponse...)
	}

	if len(bitcoincashTxsResponses) == 0 {
		return
	}

	if blockHeightInt == int64(block.Height) {
		global.NODE_LOG.Info(fmt.Sprintf("%s -> handle mempool height pending: %d", constant.GetChainName(chainId), block.Height))

		if len(bitcoincashTxsResponses) > 0 {
			for _, transaction := range bitcoincashTxsResponses {

				global.NODE_LOG.Info(fmt.Sprintf("%s -> handle mempool tx pending: %s", constant.GetChainName(chainId), transaction.TxId))

				if len(transaction.Vin) == 0 || len(transaction.Vout) == 0 {
					continue
				}

			outerCurrentTxLoop:
				for i := 0; i < len(*publicKey); i++ {
					isMonitorTx := false

					if len(transaction.Vin) > 0 {
						for _, input := range transaction.Vin {
							if strings.EqualFold((*publicKey)[i], input.Prevout.Scriptpubkey_address) {
								isMonitorTx = true
								break
							}
						}
					}

					if len(transaction.Vout) > 0 {
						for _, output := range transaction.Vout {
							if strings.EqualFold((*publicKey)[i], output.Scriptpubkey_address) {
								isMonitorTx = true
								break
							}
						}
					}

					if isMonitorTx {

						// Determine duplicate transactions
						redisTxs, err := global.NODE_REDIS.LRange(context.Background(), constantPendingTransaction, 0, -1).Result()
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							return
						}

						for _, redisTx := range redisTxs {
							if redisTx == transaction.TxId {
								continue outerCurrentTxLoop
							}
						}

						_, err = global.NODE_REDIS.RPush(context.Background(), constantPendingTransaction, transaction.TxId).Result()
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
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same sweepBlockHeight and blockHeight: %d - %d", blockHeightInt, int64(block.Height))))
	}
}
