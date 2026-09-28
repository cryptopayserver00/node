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
	sweepUtils "node/sweep/utils"
	"node/utils"
	NODE_Client "node/utils/http"
	"node/utils/notification"
	"slices"
	"strings"
	"time"
)

func GetBchBlockHeightByMempool(ctx context.Context, client NODE_Client.Client, chainId uint) int64 {
	var err error
	client.URL = constant.MempoolGetBlockHeightByNetwork(chainId)
	var blockHeight int64
	err = client.HTTPGetUnique(ctx, &blockHeight)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return 0
	}

	return blockHeight
}

func HandleBchBlockTransactionsByMempool(
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

	var bitcoincashTxsResponses []mempool.MempoolTx

	for i := 0; i < block.TxCount; i += 25 {
		client.URL = fmt.Sprintf(constant.MempoolGetBlockTransactionByNetwork(chainId), blockHash, i)
		var bitcoincashTxsResponse []mempool.MempoolTx
		err = client.HTTPGet(ctx, &bitcoincashTxsResponse)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			time.Sleep(2 * time.Second)
			return err
		}

		bitcoincashTxsResponses = append(bitcoincashTxsResponses, bitcoincashTxsResponse...)
	}

	if len(bitcoincashTxsResponses) == 0 {
		time.Sleep(2 * time.Second)
		return errors.New("not support")
	}

	if sweepBlockHeight == int64(block.Height) {
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

func HandleBchTransactionDetailsByMempool(
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

	var bitcoincashTxResponse mempool.MempoolTx
	err = client.HTTPGet(ctx, &bitcoincashTxResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	var notifyRequest request.NotificationRequest

	notifyRequest.Hash = bitcoincashTxResponse.TxId
	notifyRequest.Chain = chainId
	notifyRequest.BlockTimestamp = bitcoincashTxResponse.Status.BlockTime * 1000

	if len(bitcoincashTxResponse.Vin) == 0 || len(bitcoincashTxResponse.Vout) == 0 {
		return errors.New("not support")
	}

	_, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, "")
	if decimals == 0 {
		return errors.New("not support")
	}

	isProcess := false

	for _, input := range bitcoincashTxResponse.Vin {
		if input.Prevout.Scriptpubkey_address != "" {

			notifyRequest.FromAddress = input.Prevout.Scriptpubkey_address
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

func HandleBchPendingBlockByMempool(
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

	var bitcoincashTxsResponses []mempool.MempoolTx

	for i := 0; i < block.TxCount; i += 25 {
		client.URL = fmt.Sprintf(constant.MempoolGetBlockTransactionByNetwork(chainId), blockHash, i)
		var bitcoinTxsResponse []mempool.MempoolTx
		err = client.HTTPGet(ctx, &bitcoinTxsResponse)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			time.Sleep(2 * time.Second)
			return err
		}

		bitcoincashTxsResponses = append(bitcoincashTxsResponses, bitcoinTxsResponse...)
	}

	if len(bitcoincashTxsResponses) == 0 {
		time.Sleep(2 * time.Second)
		return errors.New("not support")
	}

	if blockHeightInt == int64(block.Height) {
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
