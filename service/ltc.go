package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
	"node/model/node/response/mempool"
	"node/model/node/response/tatum"
	"node/utils"
	"sort"
	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
)

func (n *NService) GetLtcBalance(req request.GetLtcBalance) (response.ClientBalanceResponse, error) {
	var err error
	var result response.ClientBalanceResponse

	// tatum
	client.URL = constant.TatumGetLitecoinBalance + req.Address
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(req.ChainId),
	}
	var balanceResponse tatum.LitecoinBalance
	err = client.HTTPGet(&balanceResponse)
	if err == nil {
		result.Balance, err = utils.CalSubForBtcValue(balanceResponse.Incoming, balanceResponse.Outgoing)
		if err == nil {
			return result, err
		}

		global.NODE_LOG.Error(err.Error())
	} else {
		global.NODE_LOG.Error(err.Error())
	}

	//mempool
	client.URL = fmt.Sprintf(constant.MempoolGetUtxoByNetwork(req.ChainId), req.Address)
	var utxoResponse []mempool.MempoolUtxo
	err = client.HTTPGet(&utxoResponse)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return result, err
	}

	var availableValue int64
	for _, v := range utxoResponse {
		if v.Status.Confirmed {
			availableValue += v.Value
		}
	}

	result.Balance, err = utils.FormatToBtcValue(availableValue)
	if err == nil {
		return result, nil
	}

	return result, nil
}

func (n *NService) GetLtcFeeRate(req request.GetLtcFeeRate) (tatum.LitecoinFeeRate, error) {
	var err error

	var rateResponse tatum.LitecoinFeeRate

	// tatum
	client.URL = constant.TatumGetLitecoinFeeRate
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(req.ChainId),
	}
	err = client.HTTPGet(&rateResponse)
	if err == nil {
		return rateResponse, err
	} else {
		global.NODE_LOG.Error(err.Error())
	}

	// mempool
	client.URL = constant.MempoolGetFeesyNetwork(req.ChainId)
	var feesResponse mempool.MempoolFees
	err = client.HTTPGet(&feesResponse)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return rateResponse, err
	}

	rateResponse.Fast = float64(feesResponse.FastestFee)
	rateResponse.Medium = float64(feesResponse.HalfHourFee)
	rateResponse.Slow = float64(feesResponse.HourFee)

	return rateResponse, nil
}

func (n *NService) GetLtcAddressUtxo(req request.GetLtcAddressUtxo) ([]mempool.MempoolUtxo, error) {
	var err error
	var utxos []mempool.MempoolUtxo
	var chainString string

	// tatum
	switch req.ChainId {
	case constant.LTC_MAINNET:
		chainString = "litecoin"
	case constant.LTC_TESTNET:
		chainString = "litecoin-testnet"
	default:
		return nil, errors.New("the network does not support")
	}

	totalVal := 9999999

	client.URL = fmt.Sprintf("%s?address=%s&chain=%s&totalValue=%d", constant.TatumGetLitecoinUtxo, req.Address, chainString, totalVal)
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(req.ChainId),
	}

	var tatumUtxos []tatum.LitecoinUtxo
	err = client.HTTPGet(&tatumUtxos)
	if err == nil {
		for _, v := range tatumUtxos {
			var mem mempool.MempoolUtxo
			mem.TxId = v.TxHash
			mem.Vout = v.Index
			mem.Value = utils.FormatToSatoshiValue(v.Value)
			utxos = append(utxos, mem)
		}
		return utxos, err
	} else {
		global.NODE_LOG.Error(err.Error())
	}

	// mempool
	client.URL = fmt.Sprintf(constant.MempoolGetUtxoByNetwork(req.ChainId), req.Address)
	err = client.HTTPGet(&utxos)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return utxos, err
	}

	return utxos, nil
}

func (n *NService) PostLtcBroadcast(req request.PostLtcBroadcast) (tatum.LitecoinBroadcast, error) {
	var broadcastResponse tatum.LitecoinBroadcast

	client.URL = constant.TatumGetLitecoinBroadcast
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(req.ChainId),
	}
	tatumPayload := map[string]any{
		"txData": req.TxData,
	}
	err := client.HTTPPost(tatumPayload, &broadcastResponse)
	if err == nil {
		return broadcastResponse, err
	} else {
		global.NODE_LOG.Error(err.Error())
	}

	client.URL = constant.MempoolBroadcastByNetwork(req.ChainId)
	mempoolPayload := map[string]any{
		"txHash": req.TxData,
	}
	mempoolErr := client.HTTPPost(mempoolPayload, &broadcastResponse)
	if mempoolErr != nil {
		global.NODE_LOG.Error(mempoolErr.Error())
		return broadcastResponse, mempoolErr
	}

	return broadcastResponse, nil
}

func (n *NService) GetLtcTransactions(req request.GetLtcTransactions) ([]response.ClientBtcTxResponse, error) {

	if err := n.UpdateLtcTransactionsForTatum(req); err != nil {
		global.NODE_LOG.Error(err.Error())
	}

	var txs []response.ClientBtcTxResponse

	item, err := global.NODE_MEMCACHE.Get(fmt.Sprintf(constantTransactionHistory, req.ChainId, req.Address))
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return nil, nil
	}

	err = json.Unmarshal(item.Value, &txs)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return nil, err
	}

	return txs, nil
}

func (n *NService) UpdateLtcTransactionsForTatum(req request.GetLtcTransactions) (err error) {
	var (
		saveTxs  []response.ClientBtcTxResponse
		pageSize = 50
	)

	client.URL = fmt.Sprintf("%s%s?pageSize=%d", constant.TatumGetLitecoinTransactions, req.Address, pageSize)
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(req.ChainId),
	}

	var txs []tatum.TatumLitecoinTx
	err = client.HTTPGet(&txs)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	if len(txs) == 0 {
		return
	}

	for _, v := range txs {
		tx, decodeErr := n.DecodeLtcTransactionForTatum(req.ChainId, req.Address, v)
		if decodeErr != nil {
			global.NODE_LOG.Error(decodeErr.Error())
			continue
		}
		saveTxs = append(saveTxs, tx)
	}

	sort.Slice(saveTxs, func(i, j int) bool {
		if saveTxs[i].BlockTimestamp == 0 && saveTxs[j].BlockTimestamp != 0 {
			return true
		} else if saveTxs[i].BlockTimestamp != 0 && saveTxs[j].BlockTimestamp == 0 {
			return false
		} else {
			return saveTxs[i].BlockTimestamp > saveTxs[j].BlockTimestamp
		}
	})

	byteTxs, err := json.Marshal(saveTxs)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	err = global.NODE_MEMCACHE.Set(&memcache.Item{
		Key:        fmt.Sprintf(constantTransactionHistory, req.ChainId, req.Address),
		Value:      byteTxs,
		Expiration: 86400,
	})
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	return nil
}

func (n *NService) DecodeLtcTransactionForTatum(chainId uint, address string, tx tatum.TatumLitecoinTx) (result response.ClientBtcTxResponse, err error) {
	var (
		inputs         []response.ClientBtcTxInputs
		outputs        []response.ClientBtcTxOutputs
		status         string
		blockTimestamp int64
		amount         float64
		url            string
		fee            string
		txType         string
	)

	for _, input := range tx.Inputs {
		inputFloat64, _ := strconv.ParseFloat(input.Coin.Value, 64)

		inputs = append(inputs, response.ClientBtcTxInputs{
			Address: input.Coin.Address,
			Amount:  inputFloat64,
		})
	}

	for _, output := range tx.Outputs {
		outputFloat64, _ := strconv.ParseFloat(output.Value, 64)

		outputs = append(outputs, response.ClientBtcTxOutputs{
			Address: output.Address,
			Amount:  outputFloat64,
		})
	}

	if tx.BlockNumber > 0 {
		status = "Success"
	} else {
		status = "Pending"
	}

	if tx.Time != 0 {
		blockTimestamp = int64(tx.Time * 1000)
	} else {
		tx.Time = 0
	}

	if address != "" {
		for _, v := range tx.Inputs {
			if v.Coin.Address == address {
				txType = "Send"
				break
			}
		}

		if txType == "" {
			txType = "Received"
		}
	}

	for _, v := range tx.Outputs {
		if address != "" {
			if (v.Address == address && txType == "Received") || (v.Address != address && txType == "Send") {
				value, err := strconv.ParseFloat(v.Value, 64)
				if err != nil {
					return result, err
				}
				amount += value
			}
		} else {
			if v.Address != inputs[0].Address {
				value, err := strconv.ParseFloat(v.Value, 64)
				if err != nil {
					return result, err
				}
				amount = value
				break
			}
		}
	}

	url = fmt.Sprintf("%s/%s", constant.GetSochainTxUrlByNetwork(chainId), tx.Hash)
	fee = tx.Fee

	result.Hash = tx.Hash
	result.Amount = amount
	result.Status = status
	result.BlockTimestamp = blockTimestamp
	result.Inputs = inputs
	result.Outputs = outputs
	result.Url = url
	result.Fee = fee
	result.Asset = "LTC"
	result.Type = txType

	return result, nil
}

func (n *NService) DecodeLtcHashMultiplePlatform(chainId uint, hash string) (result response.ClientBtcTxResponse, err error) {
	// tatum
	client.URL = fmt.Sprintf("%s%s", constant.TatumGetLitecoinTxByHash, hash)
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(chainId),
	}

	var txs tatum.TatumLitecoinTx
	err = client.HTTPGet(&txs)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return result, err
	}

	result, err = n.DecodeLtcTransactionForTatum(chainId, "", txs)
	if err == nil {
		return result, nil
	}

	global.NODE_LOG.Error(err.Error())

	return
}

func (n *NService) GetLtcTxByHash(req request.GetLtcTxByHash) (result response.ClientBtcTxResponse, err error) {
	item, err := global.NODE_MEMCACHE.Get(fmt.Sprintf(constantOneTransactionHistory, req.ChainId, req.Hash))
	if err == nil {
		err = json.Unmarshal(item.Value, &result)
		if err == nil && result.Status == "Success" {
			return result, nil
		}
	} else {
		global.NODE_LOG.Error(err.Error())
	}

	result, err = n.DecodeLtcHashMultiplePlatform(req.ChainId, req.Hash)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	byteTx, err := json.Marshal(result)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return result, nil
	}

	err = global.NODE_MEMCACHE.Set(&memcache.Item{
		Key:        fmt.Sprintf(constantOneTransactionHistory, req.ChainId, req.Hash),
		Value:      byteTx,
		Expiration: 86400,
	})
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return result, nil
	}

	return result, nil
}
