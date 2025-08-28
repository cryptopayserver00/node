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
	"node/utils"
	"sort"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

func (n *NService) GetBscTransactions(req request.GetBscTransactions) ([]response.ClientTransaction, error) {
	if err := n.UpdateBscTransactionsByAlchemy(req); err != nil {
		global.NODE_LOG.Error(err.Error())
	}

	item, err := global.NODE_MEMCACHE.Get(fmt.Sprintf(constantTransactionHistory, req.ChainId, req.Address))
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return nil, nil
	}

	var txs, filterTxs []response.ClientTransaction

	err = json.Unmarshal(item.Value, &txs)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return nil, err
	}

	for _, v := range txs {
		if req.Asset == "" {
			filterTxs = append(filterTxs, v)
		} else if v.Asset == req.Asset {
			filterTxs = append(filterTxs, v)
		}
	}

	sort.Slice(filterTxs, func(i, j int) bool {
		if filterTxs[i].BlockTimestamp == 0 && filterTxs[j].BlockTimestamp != 0 {
			return true
		} else if filterTxs[i].BlockTimestamp != 0 && filterTxs[j].BlockTimestamp == 0 {
			return false
		} else {
			return filterTxs[i].BlockTimestamp > filterTxs[j].BlockTimestamp
		}
	})

	return filterTxs, nil
}

func (n *NService) UpdateBscTransactionsByAlchemy(req request.GetBscTransactions) (err error) {
	client.URL = constant.GetAlchemyRPCUrlByNetwork(req.ChainId)

	var fromRpcAlchemyTxs response.RPCAlchemyTransactionDetails
	var toRpcAlchemyTxs response.RPCAlchemyTransactionDetails
	var transfers []response.RPCAlchemyTransactionTransfer
	fromPayload := map[string]any{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "alchemy_getAssetTransfers",
		"params": []map[string]any{
			{
				"fromBlock":        "0x0",
				"toBlock":          "latest",
				"fromAddress":      req.Address,
				"category":         []string{"erc20", "external"},
				"order":            "desc",
				"withMetadata":     true,
				"excludeZeroValue": true,
				"maxCount":         "0x78",
			},
		},
	}

	toPayload := map[string]any{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "alchemy_getAssetTransfers",
		"params": []map[string]any{
			{
				"fromBlock":        "0x0",
				"toBlock":          "latest",
				"toAddress":        req.Address,
				"category":         []string{"erc20", "external"},
				"order":            "desc",
				"withMetadata":     true,
				"excludeZeroValue": true,
				"maxCount":         "0x78",
			},
		},
	}

	err = client.HTTPPost(fromPayload, &fromRpcAlchemyTxs)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	err = client.HTTPPost(toPayload, &toRpcAlchemyTxs)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	transfers = append(transfers, fromRpcAlchemyTxs.Result.Transfers...)
	transfers = append(transfers, toRpcAlchemyTxs.Result.Transfers...)

	if len(transfers) == 0 {
		return
	}

	var saveTxs []response.ClientTransaction

	for _, v := range transfers {
		model, err := n.DecodeBscTransactionForAlchemy(req.ChainId, req.Address, v)
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			continue
		}
		saveTxs = append(saveTxs, model)
	}

	// get pending transaction
	// pendingTxs, err := n.GetEthPendingTransactions(req)
	// if err == nil {
	// 	saveTxs = append(saveTxs, pendingTxs...)
	// }

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

func (n *NService) DecodeBscTransactionForAlchemy(chainId uint, address string, tx response.RPCAlchemyTransactionTransfer) (model response.ClientTransaction, err error) {
	model.ChainId = chainId
	model.Address = address
	model.Hash = tx.Hash
	if tx.Asset == constant.ETH {
		model.Asset = constant.BNB

		hexString := strings.TrimPrefix(tx.RawContract.Value, "0x")
		bigIntValue, _ := new(big.Int).SetString(hexString, 16)
		model.Amount = utils.CalculateBalance(bigIntValue, 18)

	} else if tx.Category == "erc20" {
		isSupportContract, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, tx.RawContract.Address)
		if !isSupportContract {
			return
		}
		model.Asset = contractName

		hexString := strings.TrimPrefix(tx.RawContract.Value, "0x")
		bigIntValue, _ := new(big.Int).SetString(hexString, 16)
		model.Amount = utils.CalculateBalance(bigIntValue, decimals)
	} else {
		global.NODE_LOG.Error(tx.Hash)
		return
	}

	_, contractName, _, _ := sweepUtils.GetContractInfo(chainId, tx.To)

	if strings.EqualFold(tx.From, address) && contractName == constant.SWAP {
		model.Type = "Swap"
	} else if strings.EqualFold(tx.From, address) {
		model.Type = "Send"
	} else if strings.EqualFold(tx.To, address) {
		model.Type = "Received"
	} else {
		global.NODE_LOG.Error(tx.Hash)
		return
	}

	model.Status = "Success"
	model.Category = tx.Category
	timestamp, _ := time.Parse(time.RFC3339, tx.Metadata.BlockTimestamp)
	model.BlockTimestamp = int(timestamp.Unix()) * 1000

	return
}
