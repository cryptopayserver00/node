package service

import (
	"encoding/json"
	"fmt"
	"math/big"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
	"node/model/node/response/blockstream"
	"node/model/node/response/mempool"
	"node/model/node/response/tatum"
	"node/utils"
	"sort"
	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
)

func (n *NService) GetBtcBalance(req request.GetBtcBalance) (response.ClientBalanceResponse, error) {
	var err error
	var result response.ClientBalanceResponse

	// tatum
	client.URL = constant.TatumGetBitcoinBalance + req.Address
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(req.ChainId),
	}
	var balanceResponse tatum.BitcoinBalance
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

	// mempool
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

func (n *NService) GetBtcFeeRate(req request.GetBtcFeeRate) (tatum.BitcoinFeeRate, error) {
	var err error

	var rateResponse tatum.BitcoinFeeRate

	// tatum
	client.URL = constant.TatumGetBitcoinFeeRate
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

func (n *NService) GetBtcAddressUtxo(req request.GetBtcAddressUtxo) ([]blockstream.UtxoResponse, error) {
	var err error

	//blockstream
	client.URL = fmt.Sprintf("%s/address/%s/utxo", constant.GetBlcokStreamHttpUrlByNetwork(req.ChainId), req.Address)
	var utxoResponse []blockstream.UtxoResponse
	err = client.HTTPGet(&utxoResponse)
	if err == nil {
		return utxoResponse, err
	} else {
		global.NODE_LOG.Error(err.Error())
	}

	// mempool
	client.URL = fmt.Sprintf(constant.MempoolGetUtxoByNetwork(req.ChainId), req.Address)
	err = client.HTTPGet(&utxoResponse)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return utxoResponse, err
	}

	return utxoResponse, nil
}

func (n *NService) PostBtcBroadcast(req request.PostBtcBroadcast) (any, error) {
	var err error

	// tatum
	client.URL = constant.TatumGetBitcoinBroadcast
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(req.ChainId),
	}

	var tatumBroadcastResponse tatum.BitcoinBroadcast
	payload := map[string]any{
		"txData": req.TxData,
	}
	err = client.HTTPPost(payload, &tatumBroadcastResponse)
	if err == nil {
		return tatumBroadcastResponse, nil
	} else {
		global.NODE_LOG.Error(err.Error())
	}

	client.URL = fmt.Sprintf("%s/tx", constant.GetBlcokStreamHttpUrlByNetwork(req.ChainId))
	payload = map[string]any{
		"hexTransaction": req.TxData,
	}

	// blockstream
	var blockstreambroadcastResponse blockstream.BroadcastResponse
	err = client.HTTPPost(payload, &blockstreambroadcastResponse)

	if err == nil {
		return blockstreambroadcastResponse, nil
	} else {
		global.NODE_LOG.Error(err.Error())
		return nil, err
	}

}

func (n *NService) GetBtcTransactions(req request.GetBtcTransactions) ([]response.ClientBtcTxResponse, error) {

	if err := n.UpdateBtcTransactionsForBlockStream(req.ChainId, req.Address); err != nil {
		global.NODE_LOG.Error(err.Error())

		if err := n.UpdateBtcTransactionsForTatum(req.ChainId, req.Address); err != nil {
			global.NODE_LOG.Error(err.Error())
		}
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

func (n *NService) UpdateBtcTransactionsForBlockStream(chainId uint, address string) (err error) {
	var (
		saveTxs []response.ClientBtcTxResponse
	)

	client.URL = fmt.Sprintf("%s/address/%s/txs", constant.GetBlcokStreamHttpUrlByNetwork(chainId), address)

	var transactionResponse []blockstream.TransactionResponse
	err = client.HTTPGet(&transactionResponse)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	if len(transactionResponse) == 0 {
		return
	}

	for _, v := range transactionResponse {
		tx, decodeErr := n.DecodeBtcTransactionForBlockStream(chainId, address, v)
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
		Key:        fmt.Sprintf(constantTransactionHistory, chainId, address),
		Value:      byteTxs,
		Expiration: 86400,
	})
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	return nil
}

func (n *NService) UpdateBtcTransactionsForTatum(chainId uint, adress string) (err error) {
	var (
		saveTxs  []response.ClientBtcTxResponse
		pageSize = 50
	)

	client.URL = fmt.Sprintf("%s%s?pageSize=%d", constant.TatumGetBitcoinTransactions, adress, pageSize)
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(chainId),
	}

	var txs []tatum.TatumBitcoinTx
	err = client.HTTPGet(&txs)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	if len(txs) == 0 {
		return
	}

	for _, v := range txs {
		tx, decodeErr := n.DecodeBtcTransactionForTatum(chainId, adress, v)
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
		Key:        fmt.Sprintf(constantTransactionHistory, chainId, adress),
		Value:      byteTxs,
		Expiration: 86400,
	})
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	return nil
}

func (n *NService) GetBtcTransactionDetail(req request.GetBtcTransactionDetail) (result response.ClientBtcTxResponse, err error) {
	item, err := global.NODE_MEMCACHE.Get(fmt.Sprintf(constantOneTransactionHistory, req.ChainId, req.Hash))
	if err == nil {
		err = json.Unmarshal(item.Value, &result)
		if err == nil && result.Status == "Success" {
			return result, nil
		}
	} else {
		global.NODE_LOG.Error(err.Error())
	}

	result, err = n.DecodeBtcHashMultiplePlatform(req.ChainId, req.Hash)
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

func (n *NService) DecodeBtcHashMultiplePlatform(chainId uint, hash string) (result response.ClientBtcTxResponse, err error) {

	// blockstream
	client.URL = fmt.Sprintf("%s/tx/%s", constant.GetBlcokStreamHttpUrlByNetwork(chainId), hash)
	var transactionResponse blockstream.TransactionResponse
	err = client.HTTPGet(&transactionResponse)
	if err == nil {
		result, err = n.DecodeBtcTransactionForBlockStream(chainId, "", transactionResponse)
		if err == nil {
			return result, nil
		}
	}
	global.NODE_LOG.Error(err.Error())

	// tatum
	client.URL = fmt.Sprintf("%s%s", constant.TatumGetBitcoinTxByHash, hash)
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(chainId),
	}
	var txs tatum.TatumBitcoinTx
	err = client.HTTPGet(&txs)
	if err == nil {
		result, err = n.DecodeBtcTransactionForTatum(chainId, "", txs)
		if err == nil {
			return result, nil
		}
	}

	global.NODE_LOG.Error(err.Error())

	return
}

func (n *NService) DecodeBtcTransactionForBlockStream(chainId uint, address string, tx blockstream.TransactionResponse) (result response.ClientBtcTxResponse, err error) {
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

	for _, input := range tx.Vin {
		inputs = append(inputs, response.ClientBtcTxInputs{
			Address: input.Prevout.ScriptpubkeyAddress,
			Amount:  float64(input.Prevout.Value) / 100000000,
		})
	}

	for _, output := range tx.Vout {
		outputs = append(outputs, response.ClientBtcTxOutputs{
			Address: output.ScriptpubkeyAddress,
			Amount:  float64(output.Value) / 100000000,
		})
	}

	if tx.Status.Confirmed {
		status = "Success"
	} else {
		status = "Pending"
	}

	if tx.Status.BlockTime != 0 {
		blockTimestamp = int64(tx.Status.BlockTime * 1000)
	} else {
		blockTimestamp = 0
	}

	url = fmt.Sprintf("%s/%s", constant.GetBlockStreamWebsiteTxUrlByNetwork(chainId), tx.Txid)
	fee = fmt.Sprintf("%.8f", float64(tx.Fee)/100000000)

	if address != "" {
		for _, v := range tx.Vin {
			if v.Prevout.ScriptpubkeyAddress == address {
				txType = "Send"
				break
			} else {
				txType = "Received"
			}
		}
	}

	for _, output := range tx.Vout {
		if address != "" {
			if (output.ScriptpubkeyAddress == address && txType == "Received") || (output.ScriptpubkeyAddress != address && txType == "Send") {
				amount += float64(output.Value) / 100000000
			}
		} else {
			if output.ScriptpubkeyAddress != inputs[0].Address {
				amount += float64(output.Value) / 100000000
			}
		}
	}

	result.Hash = tx.Txid
	result.Amount = amount
	result.Status = status
	result.BlockTimestamp = blockTimestamp
	result.Inputs = inputs
	result.Outputs = outputs
	result.Url = url
	result.Fee = fee
	result.Asset = "BTC"
	result.Type = txType

	return result, nil
}

func (n *NService) DecodeBtcTransactionForTatum(chainId uint, address string, tx tatum.TatumBitcoinTx) (result response.ClientBtcTxResponse, err error) {
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
		inputFloat64, _ := strconv.ParseFloat(utils.CalculateBalance(big.NewInt(int64(input.Coin.Value)), 8), 64)

		inputs = append(inputs, response.ClientBtcTxInputs{
			Address: input.Coin.Address,
			Amount:  inputFloat64,
		})
	}

	for _, output := range tx.Outputs {
		outputFloat64, _ := strconv.ParseFloat(utils.CalculateBalance(big.NewInt(int64(output.Value)), 8), 64)

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
		blockTimestamp = 0
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
				value, err := strconv.ParseFloat(utils.CalculateBalance(big.NewInt(int64(v.Value)), 8), 64)
				if err != nil {
					return result, err
				}
				amount += value
			}
		} else {
			if v.Address != inputs[0].Address {
				value, err := strconv.ParseFloat(utils.CalculateBalance(big.NewInt(int64(v.Value)), 8), 64)
				if err != nil {
					return result, err
				}
				amount = value
				break
			}
		}
	}

	url = fmt.Sprintf("%s/%s", constant.GetBlockStreamWebsiteTxUrlByNetwork(chainId), tx.Hash)
	fee = utils.CalculateBalance(big.NewInt(int64(tx.Fee)), 8)

	result.Hash = tx.Hash
	result.Amount = amount
	result.Status = status
	result.BlockTimestamp = blockTimestamp
	result.Inputs = inputs
	result.Outputs = outputs
	result.Url = url
	result.Fee = fee
	result.Asset = "BTC"
	result.Type = txType

	return result, nil
}
