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
	"strings"
	"time"
)

func GetBtcBlockHeightByTatum(ctx context.Context, client NODE_Client.Client, chainId uint) int64 {
	var err error
	client.URL = constant.TatumGetBitcoinInfo
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(chainId),
	}

	var bitcoinInfoResponse tatum.TatumGetBitcoinInfo
	err = client.HTTPGet(ctx, &bitcoinInfoResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return 0
	}

	return int64(bitcoinInfoResponse.Blocks)
}

func GetBtcBlockHeightByMempool(ctx context.Context, client NODE_Client.Client, chainId uint) int64 {
	var err error
	client.URL = constant.MempoolGetBlockHeightByNetwork(chainId)
	var bitcoinHeight int64
	err = client.HTTPGetUnique(ctx, &bitcoinHeight)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return 0
	}

	return bitcoinHeight
}

func HandleBtcBlockTransactionsByTatum(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepBlockHeight int64,
	constantPendingTransaction string,
) error {
	var err error

	client.URL = constant.TatumGetBitcoinBlockByHashOrHeight + fmt.Sprint(sweepBlockHeight)
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(chainId),
	}

	var bitcoinBlockResponse tatum.TatumGetBitcoinBlock
	err = client.HTTPGet(ctx, &bitcoinBlockResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	if sweepBlockHeight == int64(bitcoinBlockResponse.Height) {
		if len(bitcoinBlockResponse.Txs) > 0 {
			for _, transaction := range bitcoinBlockResponse.Txs {

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
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", sweepBlockHeight, int64(bitcoinBlockResponse.Height))))
		time.Sleep(2 * time.Second)
		return errors.New("not the same height of block")
	}
}

func HandleBtcBlockTransactionsByMempool(
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

	var bitcoinTxsResponses []mempool.MempoolTx

	for i := 0; i < block.TxCount; i += 25 {
		client.URL = fmt.Sprintf(constant.MempoolGetBlockTransactionByNetwork(chainId), blockHash, i)
		var bitcoinTxsResponse []mempool.MempoolTx
		err = client.HTTPGet(ctx, &bitcoinTxsResponse)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			time.Sleep(2 * time.Second)
			return err
		}

		bitcoinTxsResponses = append(bitcoinTxsResponses, bitcoinTxsResponse...)
	}

	if len(bitcoinTxsResponses) == 0 {
		time.Sleep(2 * time.Second)
		return errors.New("not support")
	}

	if sweepBlockHeight == int64(block.Height) {
		if len(bitcoinTxsResponses) > 0 {
			for _, transaction := range bitcoinTxsResponses {

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

func HandleBtcTransactionDetailsByTatum(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	constantPendingTransaction string,
	txHash string,
) error {
	global.NODE_LOG.Info(fmt.Sprintf("%s -> handle tatum detail: %s", constant.GetChainName(chainId), txHash))

	var err error

	client.URL = constant.TatumGetBitcoinTxByHash + txHash
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(chainId),
	}

	var bitcoinTxResponse tatum.TatumBitcoinTx
	err = client.HTTPGet(ctx, &bitcoinTxResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	var notifyRequest request.NotificationRequest
	notifyRequest.Hash = bitcoinTxResponse.Hash
	notifyRequest.Chain = chainId
	notifyRequest.BlockTimestamp = bitcoinTxResponse.Time * 1000

	if len(bitcoinTxResponse.Inputs) == 0 || len(bitcoinTxResponse.Outputs) == 0 {
		return errors.New("not support")
	}

	_, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, "")
	if decimals == 0 {
		return errors.New("not support")
	}

	isProcess := false

	for _, input := range bitcoinTxResponse.Inputs {
		if input.Coin.Address != "" {

			notifyRequest.FromAddress = input.Coin.Address
			notifyRequest.Token = contractName

			for _, output := range bitcoinTxResponse.Outputs {
				if strings.EqualFold(output.Address, notifyRequest.FromAddress) {
					continue
				}

				notifyRequest.Amount = utils.CalculateBalance(big.NewInt(int64(output.Value)), decimals)
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

func HandleBtcTransactionDetailsByMempool(
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

	var bitcoinTxResponse mempool.MempoolTx
	err = client.HTTPGet(ctx, &bitcoinTxResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return err
	}

	var notifyRequest request.NotificationRequest

	notifyRequest.Hash = bitcoinTxResponse.TxId
	notifyRequest.Chain = chainId
	notifyRequest.BlockTimestamp = bitcoinTxResponse.Status.BlockTime * 1000

	if len(bitcoinTxResponse.Vin) == 0 || len(bitcoinTxResponse.Vout) == 0 {
		return errors.New("not support")
	}

	_, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, "")
	if decimals == 0 {
		return errors.New("not support")
	}

	isProcess := false

	for _, input := range bitcoinTxResponse.Vin {
		if input.Prevout.Scriptpubkey_address != "" {

			notifyRequest.FromAddress = input.Prevout.Scriptpubkey_address
			notifyRequest.Token = contractName

			for _, output := range bitcoinTxResponse.Vout {
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

func HandleBtcPendingBlockByTatum(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	constantPendingBlock, constantPendingTransaction string,
	blockHeight string,
	blockHeightInt int64,
) error {
	var err error

	client.URL = constant.TatumGetBitcoinBlockByHashOrHeight + blockHeight
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(chainId),
	}

	var bitcoinBlockResponse tatum.TatumGetBitcoinBlock
	err = client.HTTPGet(ctx, &bitcoinBlockResponse)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return err
	}

	if int(blockHeightInt) == bitcoinBlockResponse.Height {
		if len(bitcoinBlockResponse.Txs) > 0 {
			for _, transaction := range bitcoinBlockResponse.Txs {

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
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), fmt.Sprintf("Not the same height of block: %d - %d", blockHeightInt, int64(bitcoinBlockResponse.Height))))
		time.Sleep(2 * time.Second)
		return errors.New("not the same height of block")
	}
}

func HandleBtcPendingBlockByMempool(
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

	var bitcoinTxsResponses []mempool.MempoolTx

	for i := 0; i < block.TxCount; i += 25 {
		client.URL = fmt.Sprintf(constant.MempoolGetBlockTransactionByNetwork(chainId), blockHash, i)
		var bitcoinTxsResponse []mempool.MempoolTx
		err = client.HTTPGet(ctx, &bitcoinTxsResponse)
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			time.Sleep(2 * time.Second)
			return err
		}

		bitcoinTxsResponses = append(bitcoinTxsResponses, bitcoinTxsResponse...)
	}

	if len(bitcoinTxsResponses) == 0 {
		time.Sleep(2 * time.Second)
		return errors.New("not support")
	}

	if blockHeightInt == int64(block.Height) {
		if len(bitcoinTxsResponses) > 0 {
			for _, transaction := range bitcoinTxsResponses {

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
