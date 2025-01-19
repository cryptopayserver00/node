package service

import (
	"encoding/json"
	"fmt"
	"math/big"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
	sweepUtils "node/sweep/utils"
	"node/sweep/utils/tron"
	"sort"
	"strings"

	"node/utils"

	"github.com/bradfitz/gomemcache/memcache"
)

var (
	// tx-chainId-coin(trx,trx20)-address
	constantTronTransactionHistory = "txs-%d-%s-%s"
	trx                            = "trx"
	trc20                          = "trc20"
)

func (n *NService) GetTronTransactions(req request.GetTronTransactions) ([]response.ClientTransaction, error) {

	var values []response.ClientTransaction

	var trx request.GetTrxTransactions
	trx.ChainId = req.ChainId
	trx.Address = req.Address
	// trx
	trxTxs, err := n.GetTrxTransactions(trx)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return nil, nil
	}

	values = append(values, trxTxs...)

	isSupport, coins := sweepUtils.GetCoinsByChainId(req.ChainId)
	if isSupport && len(coins) > 0 {
		for _, item := range coins {
			if item.IsMainCoin {
				continue
			}
			if item.Decimals == 0 {
				continue
			}

			var erc20 request.GetTrc20Transactions
			erc20.ChainId = req.ChainId
			erc20.Address = req.Address
			erc20.ContractAddress = item.Contract

			var trcTxs []response.ClientTransaction

			trcTxs, err = n.GetTrc20Transactions(erc20)
			if err != nil {
				global.NODE_LOG.Error(err.Error())
				continue
			}

			values = append(values, trcTxs...)
		}
	}

	sort.Slice(values, func(i, j int) bool {
		if values[i].BlockTimestamp == 0 && values[j].BlockTimestamp != 0 {
			return true
		} else if values[i].BlockTimestamp != 0 && values[j].BlockTimestamp == 0 {
			return false
		} else {
			return values[i].BlockTimestamp > values[j].BlockTimestamp
		}
	})

	return values, nil
}

func (n *NService) GetTrxTransactions(req request.GetTrxTransactions) ([]response.ClientTransaction, error) {
	if err := n.UpdateTronTransactionsByTrongrid(req); err != nil {
		global.NODE_LOG.Error(err.Error())
	}

	item, err := global.NODE_MEMCACHE.Get(fmt.Sprintf(constantTronTransactionHistory, req.ChainId, req.Address, trx))
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return nil, nil
	}

	var allTrxs []response.ClientTransaction

	err = json.Unmarshal([]byte(item.Value), &allTrxs)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return nil, err
	}

	return allTrxs, nil
}

func (n *NService) UpdateTronTransactionsByTrongrid(req request.GetTrxTransactions) (err error) {
	limit := 120
	client.URL = fmt.Sprintf("%s/v1/accounts/%s/transactions?only_from=true&limit=%d&order_by=block_timestamp,desc&search_internal=false", constant.GetHttpUrlByNetwork(req.ChainId), req.Address, limit)
	client.Headers = map[string]string{
		"TRON-PRO-API-KEY": constant.GetRandomHTTPKeyByNetwork(req.ChainId),
	}

	var transfers []response.TronGetTxResponse
	var fromTrxs, toTrxs response.TronTxResponse
	err = client.HTTPGet(&fromTrxs)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return err
	}

	// for i := range fromTrxs.Data {
	// 	fromTrxs.Data[i].TxStatus = "Send"
	// }

	client.URL = fmt.Sprintf("%s/v1/accounts/%s/transactions?only_to=true&limit=%d&order_by=block_timestamp,desc&search_internal=false", constant.GetHttpUrlByNetwork(req.ChainId), req.Address, limit)
	err = client.HTTPGet(&toTrxs)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return err
	}

	// for i := range toTrxs.Data {
	// 	toTrxs.Data[i].TxStatus = "Received"
	// }

	transfers = append(transfers, fromTrxs.Data...)
	transfers = append(transfers, toTrxs.Data...)

	if len(transfers) == 0 {
		return
	}

	var saveTxs []response.ClientTransaction

	for _, v := range transfers {
		model, err := n.DecodeTronTransaction(req.ChainId, req.Address, v)
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			continue
		}
		saveTxs = append(saveTxs, model)
	}

	byteTxs, err := json.Marshal(saveTxs)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	err = global.NODE_MEMCACHE.Set(&memcache.Item{
		Key:        fmt.Sprintf(constantTronTransactionHistory, req.ChainId, req.Address, trx),
		Value:      byteTxs,
		Expiration: 86400,
	})
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	return nil
}

func (n *NService) DecodeTronTransaction(chainId uint, address string, tx response.TronGetTxResponse) (model response.ClientTransaction, err error) {
	model.ChainId = chainId
	model.Address = address
	model.Hash = tx.TxID
	model.Amount = utils.CalculateBalance(big.NewInt(int64(tx.RawData.Contract[0].Parameter.Value.Amount)), 6)
	model.Asset = "TRX"
	model.ContractAddress = ""

	sendAddress, err := tron.FromHexAddress(tx.RawData.Contract[0].Parameter.Value.OwnerAddress)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return model, err
	}

	if strings.EqualFold(sendAddress, address) {
		model.Type = "Send"
	}

	receiveAddress, err := tron.FromHexAddress(tx.RawData.Contract[0].Parameter.Value.ToAddress)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return model, err
	}

	if strings.EqualFold(receiveAddress, address) {
		model.Type = "Received"
	}

	model.Category = "trx"
	model.Status = tx.Ret[0].ContractRet
	model.BlockTimestamp = tx.BlockTimestamp

	return
}

func (n *NService) GetTrc20Transactions(req request.GetTrc20Transactions) ([]response.ClientTransaction, error) {
	if err := n.UpdateTrc20TransactionsByTrongrid(req); err != nil {
		global.NODE_LOG.Error(err.Error())
	}

	item, err := global.NODE_MEMCACHE.Get(fmt.Sprintf(constantTronTransactionHistory, req.ChainId, req.Address, trc20))
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return nil, nil
	}

	var allTrxs []response.ClientTransaction

	err = json.Unmarshal([]byte(item.Value), &allTrxs)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return nil, err
	}

	if req.ContractAddress != "" {
		var newTxs []response.ClientTransaction
		for _, innerItem := range allTrxs {
			if strings.EqualFold(innerItem.ContractAddress, req.ContractAddress) {
				newTxs = append(newTxs, innerItem)
			}
		}

		return newTxs, nil
	}

	return allTrxs, nil
}

func (n *NService) UpdateTrc20TransactionsByTrongrid(req request.GetTrc20Transactions) (err error) {
	limit := 120
	client.URL = fmt.Sprintf("%s/v1/accounts/%s/transactions/trc20?only_from=true&limit=%d&contract_address=%s&search_internal=false", constant.GetHttpUrlByNetwork(req.ChainId), req.Address, limit, req.ContractAddress)
	client.Headers = map[string]string{
		"TRON-PRO-API-KEY": constant.GetRandomHTTPKeyByNetwork(req.ChainId),
	}

	var transfers []response.TronGetTrc20Response
	var fromTrxs, toTrxs response.TronTrc20Response
	err = client.HTTPGet(&fromTrxs)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return err
	}

	// for i := range fromTrxs.Data {
	// 	fromTrxs.Data[i].TxStatus = "Send"
	// }

	client.URL = fmt.Sprintf("%s/v1/accounts/%s/transactions/trc20?only_to=true&limit=%d&contract_address=%s&search_internal=false", constant.GetHttpUrlByNetwork(req.ChainId), req.Address, limit, req.ContractAddress)
	err = client.HTTPGet(&toTrxs)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return err
	}

	// for i := range toTrxs.Data {
	// 	toTrxs.Data[i].TxStatus = "Received"
	// }

	transfers = append(transfers, fromTrxs.Data...)
	transfers = append(transfers, toTrxs.Data...)

	if len(transfers) == 0 {
		return
	}

	var saveTxs []response.ClientTransaction

	for _, v := range transfers {
		model, err := n.DecodeTrc20Transaction(req.ChainId, req.Address, v)
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			continue
		}
		saveTxs = append(saveTxs, model)
	}

	byteTxs, err := json.Marshal(saveTxs)
	if err != nil {
		return
	}

	err = global.NODE_MEMCACHE.Set(&memcache.Item{
		Key:        fmt.Sprintf(constantTronTransactionHistory, req.ChainId, req.Address, trc20),
		Value:      byteTxs,
		Expiration: 86400,
	})
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	return nil
}

func (n *NService) DecodeTrc20Transaction(chainId uint, address string, tx response.TronGetTrc20Response) (model response.ClientTransaction, err error) {
	model.ChainId = chainId
	model.Address = address
	model.Hash = tx.TransactionId

	value, err := utils.HexStringToBigInt(tx.Value)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return model, err
	}

	model.Amount = utils.CalculateBalance(value, tx.TokenInfo.Decimals)
	model.Asset = tx.TokenInfo.Symbol
	model.ContractAddress = tx.TokenInfo.Address

	if strings.EqualFold(tx.From, address) {
		model.Type = "Send"
	}

	if strings.EqualFold(tx.To, address) {
		model.Type = "Received"
	}

	model.Category = "trc20"
	model.Status = "SUCCESS"
	model.BlockTimestamp = tx.BlockTimestamp

	return
}
