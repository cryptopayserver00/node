package plugin

import (
	"context"
	"fmt"
	"math/big"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response/mempool"
	"node/model/node/response/tatum"
	sweepUtils "node/sweep/utils"
	"node/utils"
	NODE_Client "node/utils/http"
	"node/utils/notification"
	"strconv"
	"strings"
)

func GetLtcBlockHeightByTatum(client NODE_Client.Client, chainId uint) int64 {
	var err error
	client.URL = constant.TatumGetLitecoinInfo
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(chainId),
	}

	var litecoinInfoResponse tatum.TatumGetLitecoinInfo
	err = client.HTTPGet(&litecoinInfoResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return 0
	}

	return int64(litecoinInfoResponse.Blocks)
}

func GetLtcBlockHeightByMempool(client NODE_Client.Client, chainId uint) int64 {
	var err error
	client.URL = constant.MempoolGetBlockHeightByNetwork(chainId)
	var height int64
	err = client.HTTPGetUnique(&height)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return 0
	}

	return height
}

func HandleLtcBlockTransactionsByTatum(
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepCount *map[int64]int,
	sweepBlockHeight *int64,
	constantSweepBlock, constantPendingTransaction string,
) {
	var err error
	client.URL = constant.TatumGetLitecoinBlockByHashOrHeight + fmt.Sprint(*sweepBlockHeight)
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(chainId),
	}

	var litecoinBlockResponse tatum.TatumGetLitecoinBlock
	err = client.HTTPGet(&litecoinBlockResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	if *sweepBlockHeight == int64(litecoinBlockResponse.Height) {

		if len(litecoinBlockResponse.Txs) > 0 {
			for _, transaction := range litecoinBlockResponse.Txs {

				if len(transaction.Inputs) == 0 || len(transaction.Outputs) == 0 {
					continue
				}

			outerCurrentTxLoop:
				for i := 0; i < len(*publicKey); i++ {
					isMonitorTx := false

					if len(transaction.Inputs) > 0 {
						for _, input := range transaction.Inputs {
							if strings.EqualFold((*publicKey)[i], input.Coin.Address) {
								isMonitorTx = true
								break
							}
						}
					}

					if len(transaction.Outputs) > 0 {
						for _, output := range transaction.Outputs {
							if strings.EqualFold((*publicKey)[i], output.Address) {
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
							if redisTx == transaction.Hash {
								continue outerCurrentTxLoop
							}
						}

						_, err = global.NODE_REDIS.RPush(context.Background(), constantPendingTransaction, transaction.Hash).Result()
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
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", *sweepBlockHeight, int64(litecoinBlockResponse.Height))))
	}
}

func HandleLtcBlockTransactionsByMempool(
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

	var litecoinTxsResponses []mempool.MempoolTx

	for i := 0; i < block.TxCount; i += 25 {
		client.URL = fmt.Sprintf(constant.MempoolGetBlockTransactionByNetwork(chainId), blockHash, i)
		var litecoinTxsResponse []mempool.MempoolTx
		err = client.HTTPGet(&litecoinTxsResponse)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return
		}

		litecoinTxsResponses = append(litecoinTxsResponses, litecoinTxsResponse...)
	}

	if len(litecoinTxsResponses) == 0 {
		return
	}

	if *sweepBlockHeight == int64(block.Height) {

		if len(litecoinTxsResponses) > 0 {
			for _, transaction := range litecoinTxsResponses {

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

func HandleLtcTransactionDetailsByTatum(
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	constantPendingTransaction string,
	txHash string,
) {

	global.NODE_LOG.Info(fmt.Sprintf("%s -> handle tatum detail: %s", constant.GetChainName(chainId), txHash))

	var err error
	var isProcess bool

	client.URL = constant.TatumGetLitecoinTxByHash + txHash
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(chainId),
	}

	var litecoinTxResponse tatum.TatumLitecoinTx
	err = client.HTTPGet(&litecoinTxResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	if len(litecoinTxResponse.Inputs) == 0 || len(litecoinTxResponse.Outputs) == 0 {
		return
	}

	_, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, "")
	if decimals == 0 {
		return
	}

	var notifyRequest request.NotificationRequest
	notifyRequest.Hash = litecoinTxResponse.Hash
	notifyRequest.Chain = chainId
	notifyRequest.Token = contractName

	if len(strconv.Itoa(litecoinTxResponse.Time)) == 10 {
		litecoinTxResponse.Time *= 1000
	}
	notifyRequest.BlockTimestamp = litecoinTxResponse.Time

	for _, input := range litecoinTxResponse.Inputs {
		if input.Coin.Address != "" {
			notifyRequest.FromAddress = input.Coin.Address
			continue
		}
	}

	for _, output := range litecoinTxResponse.Outputs {
		if strings.EqualFold(output.Address, notifyRequest.FromAddress) {
			continue
		}

		notifyRequest.Amount = output.Value
		for _, v := range *publicKey {
			notifyRequest.Address = v
			notifyRequest.ToAddress = output.Address

			if strings.EqualFold(notifyRequest.FromAddress, v) {
				notifyRequest.TransactType = "send"

				err = notification.NotificationRequest(notifyRequest)
				if err != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
					return
				}
				isProcess = true
			}

			if strings.EqualFold(output.Address, v) {
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

	if isProcess {
		_, err = global.NODE_REDIS.LPop(context.Background(), constantPendingTransaction).Result()
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		}
	} else {
		global.NODE_LOG.Error(fmt.Sprintf("Can not handle the tx: %s, Retry | %s -> %s", txHash, constant.GetChainName(chainId), err.Error()))
	}
}

func HandleLtcTransactionDetailsByMempool(
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

	var litecoinTxResponse mempool.MempoolTx
	err = client.HTTPGet(&litecoinTxResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	_, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, "")
	if decimals == 0 {
		return
	}

	var notifyRequest request.NotificationRequest

	notifyRequest.Hash = litecoinTxResponse.TxId
	notifyRequest.Chain = chainId
	notifyRequest.BlockTimestamp = litecoinTxResponse.Status.BlockTime * 1000
	notifyRequest.Token = contractName

	if len(litecoinTxResponse.Vin) == 0 || len(litecoinTxResponse.Vout) == 0 {
		return
	}

	for _, input := range litecoinTxResponse.Vin {
		if input.Prevout.Scriptpubkey_address != "" {
			notifyRequest.FromAddress = input.Prevout.Scriptpubkey_address
			continue
		}
	}

	for _, output := range litecoinTxResponse.Vout {
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

	if isProcess {
		_, err = global.NODE_REDIS.LPop(context.Background(), constantPendingTransaction).Result()
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		}
	} else {
		global.NODE_LOG.Error(fmt.Sprintf("Can not handle the tx: %s, Retry | %s -> %s", txHash, constant.GetChainName(chainId), err.Error()))
	}
}

func HandleLtcPendingBlockByTatum(
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	constantPendingBlock, constantPendingTransaction string,
	blockHeight string,
	blockHeightInt int64,
) {
	var err error

	client.URL = constant.TatumGetLitecoinBlockByHashOrHeight + blockHeight
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(chainId),
	}

	var litecoinBlockResponse tatum.TatumGetLitecoinBlock
	err = client.HTTPGet(&litecoinBlockResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	if int(blockHeightInt) == litecoinBlockResponse.Height {
		global.NODE_LOG.Info(fmt.Sprintf("%s -> handle tatum height pending: %s", constant.GetChainName(chainId), blockHeight))

		if len(litecoinBlockResponse.Txs) > 0 {
			for _, transaction := range litecoinBlockResponse.Txs {

				global.NODE_LOG.Info(fmt.Sprintf("%s -> handle tatum tx pending: %s", constant.GetChainName(chainId), transaction.Hash))

				if len(transaction.Inputs) == 0 || len(transaction.Outputs) == 0 {
					continue
				}

			outerCurrentTxLoop:
				for i := 0; i < len(*publicKey); i++ {
					isMonitorTx := false

					if len(transaction.Inputs) > 0 {
						for _, input := range transaction.Inputs {
							if strings.EqualFold((*publicKey)[i], input.Coin.Address) {
								isMonitorTx = true
								break
							}
						}
					}

					if len(transaction.Outputs) > 0 {
						for _, output := range transaction.Outputs {
							if strings.EqualFold((*publicKey)[i], output.Address) {
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
							if redisTx == transaction.Hash {
								continue outerCurrentTxLoop
							}
						}

						_, err = global.NODE_REDIS.RPush(context.Background(), constantPendingTransaction, transaction.Hash).Result()
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
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", blockHeightInt, int64(litecoinBlockResponse.Height))))
	}
}

func HandleLtcPendingBlockByMempool(
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

	var litecoinTxsResponses []mempool.MempoolTx

	for i := 0; i < block.TxCount; i += 25 {
		client.URL = fmt.Sprintf(constant.MempoolGetBlockTransactionByNetwork(chainId), blockHash, i)
		var litecoinTxsResponse []mempool.MempoolTx
		err = client.HTTPGet(&litecoinTxsResponse)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return
		}

		litecoinTxsResponses = append(litecoinTxsResponses, litecoinTxsResponse...)
	}

	if len(litecoinTxsResponses) == 0 {
		return
	}

	if blockHeightInt == int64(block.Height) {
		global.NODE_LOG.Info(fmt.Sprintf("%s -> handle mempool height pending: %d", constant.GetChainName(chainId), block.Height))

		if len(litecoinTxsResponses) > 0 {
			for _, transaction := range litecoinTxsResponses {

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
