package plugin

import (
	"context"
	"errors"
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
	"slices"
	"strconv"
	"strings"
	"time"
)

func GetLtcBlockHeightByTatum(ctx context.Context, client NODE_Client.Client, chainId uint) int64 {
	var err error
	client.URL = constant.TatumGetLitecoinInfo
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(chainId),
	}

	var litecoinInfoResponse tatum.TatumGetLitecoinInfo
	err = client.HTTPGet(ctx, &litecoinInfoResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return 0
	}

	return int64(litecoinInfoResponse.Blocks)
}

func GetLtcBlockHeightByMempool(ctx context.Context, client NODE_Client.Client, chainId uint) int64 {
	var err error
	client.URL = constant.MempoolGetBlockHeightByNetwork(chainId)
	var height int64
	err = client.HTTPGetUnique(ctx, &height)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return 0
	}

	return height
}

func HandleLtcBlockTransactionsByTatum(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepBlockHeight int64,
	constantPendingTransaction string,
) error {
	var err error
	client.URL = constant.TatumGetLitecoinBlockByHashOrHeight + fmt.Sprint(sweepBlockHeight)
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(chainId),
	}

	var litecoinBlockResponse tatum.TatumGetLitecoinBlock
	err = client.HTTPGet(ctx, &litecoinBlockResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	if sweepBlockHeight == int64(litecoinBlockResponse.Height) {
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
						redisTxs, err := global.NODE_REDIS.LRange(ctx, constantPendingTransaction, 0, -1).Result()
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							time.Sleep(2 * time.Second)
							return err
						}

						found := slices.Contains(redisTxs, transaction.Hash)
						if found {
							break outerCurrentTxLoop

						} else {
							_, err = global.NODE_REDIS.RPush(ctx, constantPendingTransaction, transaction.Hash).Result()
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
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", sweepBlockHeight, int64(litecoinBlockResponse.Height))))
		time.Sleep(2 * time.Second)
		return errors.New("not the same height of block")
	}
}

func HandleLtcBlockTransactionsByMempool(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepBlockHeight int64,
	constantPendingTransaction string,
) error {
	var err error

	var blockHash string
	client.URL = fmt.Sprintf(constant.MempoolGetBlockHashByNetwork(chainId), sweepBlockHeight)
	err = client.HTTPGetUnique(ctx, &blockHash)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	var block mempool.MempoolBlock
	client.URL = fmt.Sprintf(constant.MempoolGetBlockByNetwork(chainId), blockHash)
	err = client.HTTPGet(ctx, &block)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	var litecoinTxsResponses []mempool.MempoolTx

	for i := 0; i < block.TxCount; i += 25 {
		client.URL = fmt.Sprintf(constant.MempoolGetBlockTransactionByNetwork(chainId), blockHash, i)
		var litecoinTxsResponse []mempool.MempoolTx
		err = client.HTTPGet(ctx, &litecoinTxsResponse)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			time.Sleep(2 * time.Second)
			return err
		}

		litecoinTxsResponses = append(litecoinTxsResponses, litecoinTxsResponse...)
	}

	if len(litecoinTxsResponses) == 0 {
		time.Sleep(2 * time.Second)
		return errors.New("not support")
	}

	if sweepBlockHeight == int64(block.Height) {
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
						redisTxs, err := global.NODE_REDIS.LRange(ctx, constantPendingTransaction, 0, -1).Result()
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							time.Sleep(2 * time.Second)
							return err
						}

						found := slices.Contains(redisTxs, transaction.TxId)
						if found {
							break outerCurrentTxLoop
						} else {
							_, err = global.NODE_REDIS.RPush(ctx, constantPendingTransaction, transaction.TxId).Result()
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
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", sweepBlockHeight, int64(block.Height))))
		time.Sleep(2 * time.Second)
		return errors.New("not the same height of block")
	}
}

func HandleLtcTransactionDetailsByTatum(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	constantPendingTransaction string,
	txHash string,
) error {
	global.NODE_LOG.Info(fmt.Sprintf("%s -> handle tatum detail: %s", constant.GetChainName(chainId), txHash))

	var err error

	client.URL = constant.TatumGetLitecoinTxByHash + txHash
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(chainId),
	}

	var litecoinTxResponse tatum.TatumLitecoinTx
	err = client.HTTPGet(ctx, &litecoinTxResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	var notifyRequest request.NotificationRequest
	notifyRequest.Hash = litecoinTxResponse.Hash
	notifyRequest.Chain = chainId

	if len(litecoinTxResponse.Inputs) == 0 || len(litecoinTxResponse.Outputs) == 0 {
		return errors.New("not support")
	}

	_, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, "")
	if decimals == 0 {
		return errors.New("not support")
	}

	notifyRequest.Token = contractName

	if len(strconv.Itoa(litecoinTxResponse.Time)) == 10 {
		litecoinTxResponse.Time *= 1000
	}

	isProcess := false

	for _, input := range litecoinTxResponse.Inputs {
		if input.Coin.Address != "" {

			notifyRequest.FromAddress = input.Coin.Address

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

						err = notification.NotificationRequest(ctx, notifyRequest)
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							return err
						}
						isProcess = true
					}

					if strings.EqualFold(output.Address, v) {
						notifyRequest.TransactType = "receive"

						err = notification.NotificationRequest(ctx, notifyRequest)
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							return err
						}
						isProcess = true
					}
				}
			}
		}
	}

	if !isProcess {
		return fmt.Errorf("no monitored address matched for tx: %s", notifyRequest.Hash)
	}

	return nil
}

func HandleLtcTransactionDetailsByMempool(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	constantPendingTransaction string,
	txHash string,
) error {
	global.NODE_LOG.Info(fmt.Sprintf("%s -> handle mempool detail: %s", constant.GetChainName(chainId), txHash))

	var err error

	client.URL = fmt.Sprintf(constant.MempoolGetTransctionByNetwork(chainId), txHash)

	var litecoinTxResponse mempool.MempoolTx
	err = client.HTTPGet(ctx, &litecoinTxResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	var notifyRequest request.NotificationRequest

	notifyRequest.Hash = litecoinTxResponse.TxId
	notifyRequest.Chain = chainId
	notifyRequest.BlockTimestamp = litecoinTxResponse.Status.BlockTime * 1000

	if len(litecoinTxResponse.Vin) == 0 || len(litecoinTxResponse.Vout) == 0 {
		return errors.New("not support")
	}

	_, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, "")
	if decimals == 0 {
		return errors.New("not support")
	}
	notifyRequest.Token = contractName

	isProcess := false

	for _, input := range litecoinTxResponse.Vin {
		if input.Prevout.Scriptpubkey_address != "" {

			notifyRequest.FromAddress = input.Prevout.Scriptpubkey_address

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

						err = notification.NotificationRequest(ctx, notifyRequest)
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							return err
						}
						isProcess = true
					}

					if strings.EqualFold(output.Scriptpubkey_address, v) {
						notifyRequest.TransactType = "receive"

						err = notification.NotificationRequest(ctx, notifyRequest)
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							return err
						}
						isProcess = true
					}
				}

			}
		}
	}

	if !isProcess {
		return fmt.Errorf("no monitored address matched for tx: %s", notifyRequest.Hash)
	}

	return nil
}

func HandleLtcPendingBlockByTatum(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	constantPendingBlock, constantPendingTransaction string,
	blockHeight string,
	blockHeightInt int64,
) error {
	var err error

	client.URL = constant.TatumGetLitecoinBlockByHashOrHeight + blockHeight
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(chainId),
	}

	var litecoinBlockResponse tatum.TatumGetLitecoinBlock
	err = client.HTTPGet(ctx, &litecoinBlockResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return err
	}

	if int(blockHeightInt) == litecoinBlockResponse.Height {
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
						redisTxs, err := global.NODE_REDIS.LRange(ctx, constantPendingTransaction, 0, -1).Result()
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							time.Sleep(2 * time.Second)
							return err
						}

						found := slices.Contains(redisTxs, transaction.Hash)
						if found {
							break outerCurrentTxLoop

						} else {
							_, err = global.NODE_REDIS.RPush(ctx, constantPendingTransaction, transaction.Hash).Result()
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
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", blockHeightInt, int64(litecoinBlockResponse.Height))))
		time.Sleep(2 * time.Second)
		return errors.New("not the same height of block")
	}
}

func HandleLtcPendingBlockByMempool(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	constantPendingBlock, constantPendingTransaction string,
	blockHeight string,
	blockHeightInt int64,
) error {
	var err error

	var blockHash string
	client.URL = fmt.Sprintf(constant.MempoolGetBlockHashByNetwork(chainId), blockHeightInt)
	err = client.HTTPGetUnique(ctx, &blockHash)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	var block mempool.MempoolBlock
	client.URL = fmt.Sprintf(constant.MempoolGetBlockByNetwork(chainId), blockHash)
	err = client.HTTPGet(ctx, &block)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	var litecoinTxsResponses []mempool.MempoolTx

	for i := 0; i < block.TxCount; i += 25 {
		client.URL = fmt.Sprintf(constant.MempoolGetBlockTransactionByNetwork(chainId), blockHash, i)
		var litecoinTxsResponse []mempool.MempoolTx
		err = client.HTTPGet(ctx, &litecoinTxsResponse)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			time.Sleep(2 * time.Second)
			return err
		}

		litecoinTxsResponses = append(litecoinTxsResponses, litecoinTxsResponse...)
	}

	if len(litecoinTxsResponses) == 0 {
		time.Sleep(2 * time.Second)
		return errors.New("not support")
	}

	if blockHeightInt == int64(block.Height) {
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
						redisTxs, err := global.NODE_REDIS.LRange(ctx, constantPendingTransaction, 0, -1).Result()
						if err != nil {
							global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
							time.Sleep(2 * time.Second)
							return err
						}

						found := slices.Contains(redisTxs, transaction.TxId)
						if found {
							break outerCurrentTxLoop
						} else {
							_, err = global.NODE_REDIS.RPush(ctx, constantPendingTransaction, transaction.TxId).Result()
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
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same sweepBlockHeight and blockHeight: %d - %d", blockHeightInt, int64(block.Height))))
		time.Sleep(2 * time.Second)
		return errors.New("not the same height of block")
	}
}
