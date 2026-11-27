package service

import (
	"context"
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
	"github.com/gin-gonic/gin"
)

func (n *NService) GetBtcBalance(c *gin.Context, req request.GetBtcBalance) (response.ClientBalanceResponse, error) {
	ctx := c.Request.Context()

	var err error
	var result response.ClientBalanceResponse

	// tatum
	client.URL = constant.TatumGetBitcoinBalance + req.Address
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(req.ChainId),
	}
	var balanceResponse tatum.BitcoinBalance
	err = client.HTTPGet(ctx, &balanceResponse)
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
	err = client.HTTPGet(ctx, &utxoResponse)
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

func (n *NService) GetBtcFeeRate(c *gin.Context, req request.GetBtcFeeRate) (tatum.BitcoinFeeRate, error) {
	ctx := c.Request.Context()

	var err error

	var rateResponse tatum.BitcoinFeeRate

	// tatum
	client.URL = constant.TatumGetBitcoinFeeRate
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(req.ChainId),
	}

	err = client.HTTPGet(ctx, &rateResponse)
	if err == nil {
		return rateResponse, err
	} else {
		global.NODE_LOG.Error(err.Error())
	}

	// mempool
	client.URL = constant.MempoolGetFeesyNetwork(req.ChainId)
	var feesResponse mempool.MempoolFees
	err = client.HTTPGet(ctx, &feesResponse)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return rateResponse, err
	}

	rateResponse.Fast = float64(feesResponse.FastestFee)
	rateResponse.Medium = float64(feesResponse.HalfHourFee)
	rateResponse.Slow = float64(feesResponse.HourFee)

	return rateResponse, nil
}

func (n *NService) GetBtcAddressUtxo(c *gin.Context, req request.GetBtcAddressUtxo) ([]blockstream.UtxoResponse, error) {
	ctx := c.Request.Context()

	var err error

	//blockstream
	client.URL = fmt.Sprintf("%s/address/%s/utxo", constant.GetBlcokStreamHttpUrlByNetwork(req.ChainId), req.Address)
	var utxoResponse []blockstream.UtxoResponse
	err = client.HTTPGet(ctx, &utxoResponse)
	if err == nil {
		return utxoResponse, err
	} else {
		global.NODE_LOG.Error(err.Error())
	}

	// mempool
	client.URL = fmt.Sprintf(constant.MempoolGetUtxoByNetwork(req.ChainId), req.Address)
	err = client.HTTPGet(ctx, &utxoResponse)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return utxoResponse, err
	}

	return utxoResponse, nil
}

func (n *NService) PostBtcBroadcast(c *gin.Context, req request.PostBtcBroadcast) (any, error) {
	ctx := c.Request.Context()

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
	err = client.HTTPPost(ctx, payload, &tatumBroadcastResponse)
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
	err = client.HTTPPost(ctx, payload, &blockstreambroadcastResponse)

	if err == nil {
		return blockstreambroadcastResponse, nil
	} else {
		global.NODE_LOG.Error(err.Error())
		return nil, err
	}

}

func (n *NService) GetBtcTransactions(c *gin.Context, req request.GetBtcTransactions) ([]response.ClientBtcTxResponse, error) {
	ctx := c.Request.Context()

	txs, err := n.fetchBtcTransactionsLive(ctx, req.ChainId, req.Address)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		// Fall back to cache, then Tatum update path.
		if cached, cacheErr := n.readBtcTxCache(req.ChainId, req.Address); cacheErr == nil && len(cached) > 0 {
			return cached, nil
		}
		if uErr := n.UpdateBtcTransactionsForTatum(ctx, req.ChainId, req.Address); uErr != nil {
			global.NODE_LOG.Error(uErr.Error())
			return []response.ClientBtcTxResponse{}, nil
		}
		if cached, cacheErr := n.readBtcTxCache(req.ChainId, req.Address); cacheErr == nil {
			return cached, nil
		}
		return []response.ClientBtcTxResponse{}, nil
	}

	n.writeBtcTxCache(req.ChainId, req.Address, txs)
	return txs, nil
}

func (n *NService) readBtcTxCache(chainId uint, address string) ([]response.ClientBtcTxResponse, error) {
	if !memcacheEnabled() {
		return nil, fmt.Errorf("memcache disabled")
	}
	item, err := global.NODE_MEMCACHE.Get(fmt.Sprintf(constantTransactionHistory, chainId, address))
	if err != nil {
		return nil, err
	}
	var txs []response.ClientBtcTxResponse
	if err := json.Unmarshal(item.Value, &txs); err != nil {
		return nil, err
	}
	return txs, nil
}

func (n *NService) writeBtcTxCache(chainId uint, address string, txs []response.ClientBtcTxResponse) {
	if !memcacheEnabled() || len(txs) == 0 {
		return
	}
	raw, err := json.Marshal(txs)
	if err != nil {
		return
	}
	_ = global.NODE_MEMCACHE.Set(&memcache.Item{
		Key:        fmt.Sprintf(constantTransactionHistory, chainId, address),
		Value:      raw,
		Expiration: 300,
	})
}

// fetchBtcTransactionsLive pulls confirmed + mempool txs from Esplora-compatible APIs.
func (n *NService) fetchBtcTransactionsLive(ctx context.Context, chainId uint, address string) ([]response.ClientBtcTxResponse, error) {
	var raw []blockstream.TransactionResponse

	appendFrom := func(url string) error {
		if url == "" {
			return fmt.Errorf("empty url")
		}
		httpClient := client
		httpClient.URL = url
		var rows []blockstream.TransactionResponse
		if err := httpClient.HTTPGet(ctx, &rows); err != nil {
			return err
		}
		raw = append(raw, rows...)
		return nil
	}

	blockstreamURL := fmt.Sprintf("%s/address/%s/txs", constant.GetBlcokStreamHttpUrlByNetwork(chainId), address)
	mempoolTxURL := fmt.Sprintf(constant.MempoolGetAddressTxsByNetwork(chainId), address)
	mempoolPendingURL := fmt.Sprintf(constant.MempoolGetAddressMempoolTxsByNetwork(chainId), address)

	var lastErr error
	if err := appendFrom(blockstreamURL); err != nil {
		lastErr = err
		global.NODE_LOG.Error(fmt.Sprintf("btc blockstream txs: %v", err))
		if err := appendFrom(mempoolTxURL); err != nil {
			lastErr = err
			global.NODE_LOG.Error(fmt.Sprintf("btc mempool txs: %v", err))
		}
	}
	if err := appendFrom(mempoolPendingURL); err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("btc mempool pending txs: %v", err))
		if lastErr == nil {
			lastErr = err
		}
	}

	if len(raw) == 0 {
		if lastErr != nil {
			return nil, lastErr
		}
		return []response.ClientBtcTxResponse{}, nil
	}

	seen := make(map[string]response.ClientBtcTxResponse, len(raw))
	order := make([]string, 0, len(raw))
	for _, v := range raw {
		tx, decodeErr := n.DecodeBtcTransactionForBlockStream(chainId, address, v)
		if decodeErr != nil || tx.Hash == "" {
			continue
		}
		key := tx.Hash
		if existing, ok := seen[key]; ok {
			// Prefer Success over Pending.
			if existing.Status != "Success" && tx.Status == "Success" {
				seen[key] = tx
			}
			continue
		}
		seen[key] = tx
		order = append(order, key)
	}

	out := make([]response.ClientBtcTxResponse, 0, len(order))
	for _, key := range order {
		out = append(out, seen[key])
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].BlockTimestamp == 0 && out[j].BlockTimestamp != 0 {
			return true
		}
		if out[i].BlockTimestamp != 0 && out[j].BlockTimestamp == 0 {
			return false
		}
		return out[i].BlockTimestamp > out[j].BlockTimestamp
	})
	return out, nil
}

func (n *NService) UpdateBtcTransactionsForBlockStream(ctx context.Context, chainId uint, address string) (err error) {
	var (
		saveTxs []response.ClientBtcTxResponse
	)

	client.URL = fmt.Sprintf("%s/address/%s/txs", constant.GetBlcokStreamHttpUrlByNetwork(chainId), address)

	var transactionResponse []blockstream.TransactionResponse
	err = client.HTTPGet(ctx, &transactionResponse)
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

func (n *NService) UpdateBtcTransactionsForTatum(ctx context.Context, chainId uint, adress string) (err error) {
	var (
		saveTxs  []response.ClientBtcTxResponse
		pageSize = 50
	)

	client.URL = fmt.Sprintf("%s%s?pageSize=%d", constant.TatumGetBitcoinTransactions, adress, pageSize)
	client.Headers = map[string]string{
		"x-api-key": constant.GetTatumRandomKeyByNetwork(chainId),
	}

	var txs []tatum.TatumBitcoinTx
	err = client.HTTPGet(ctx, &txs)
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

func (n *NService) GetBtcTransactionDetail(c *gin.Context, req request.GetBtcTransactionDetail) (result response.ClientBtcTxResponse, err error) {
	ctx := c.Request.Context()

	item, err := global.NODE_MEMCACHE.Get(fmt.Sprintf(constantOneTransactionHistory, req.ChainId, req.Hash))
	if err == nil {
		err = json.Unmarshal(item.Value, &result)
		if err == nil && result.Status == "Success" {
			return result, nil
		}
	} else {
		global.NODE_LOG.Error(err.Error())
	}

	result, err = n.DecodeBtcHashMultiplePlatform(ctx, req.ChainId, req.Hash)
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

func (n *NService) DecodeBtcHashMultiplePlatform(ctx context.Context, chainId uint, hash string) (result response.ClientBtcTxResponse, err error) {

	// blockstream
	client.URL = fmt.Sprintf("%s/tx/%s", constant.GetBlcokStreamHttpUrlByNetwork(chainId), hash)
	var transactionResponse blockstream.TransactionResponse
	err = client.HTTPGet(ctx, &transactionResponse)
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
	err = client.HTTPGet(ctx, &txs)
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
