package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
	chainUtils "node/sweep/utils"
	sweepUtils "node/sweep/utils"
	"node/sweep/utils/erc20"
	"node/utils"
	"sort"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

var (
	// txs-chainId-address
	constantTransactionHistory = "txs-%d-%s"
	// tx-chainId-hash
	constantOneTransactionHistory = "tx-%d-%s"
	// tx-pending-chainId-hash
	constantOnePendingTransactionHistory = "tx-pending-%d-%s"
)

func (n *NService) GetEthPendingTransactions(req request.GetEthTransactions) ([]response.ClientTransaction, error) {
	var err error
	var pendings []response.ClientTransaction

	client.URL = constant.GetGeneralRPCUrlByNetwork(req.ChainId)
	payload := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "txpool_contentFrom",
		"params":  []string{req.Address},
	}
	var rpcGeneralResponse response.RPCGeneralTxpool
	err = client.HTTPPost(payload, &rpcGeneralResponse)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return nil, err
	}

	for _, txInfo := range rpcGeneralResponse.Result.Pending {
		model, err := n.DecodeEthTransactionForRpc(req.ChainId, req.Address, txInfo)
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			continue
		}

		byteTxs, err := json.Marshal(txInfo)
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			return nil, err
		}

		// save pending tx to nosql
		err = global.NODE_MEMCACHE.Set(&memcache.Item{
			Key:        fmt.Sprintf(constantOnePendingTransactionHistory, req.ChainId, txInfo.Hash),
			Value:      byteTxs,
			Expiration: 86400,
		})
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			return nil, err
		}

		pendings = append(pendings, model)
	}
	return pendings, nil
}

func (n *NService) GetEthPendingTransaction(req request.GetEthPendingTransaction) (response.RPCTransaction, error) {
	var tx response.RPCTransaction

	item, err := global.NODE_MEMCACHE.Get(fmt.Sprintf(constantOnePendingTransactionHistory, req.ChainId, req.Hash))
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return tx, nil
	}

	err = json.Unmarshal(item.Value, &tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return tx, err
	}

	return tx, nil
}

func (n *NService) GetEthTransactions(req request.GetEthTransactions) ([]response.ClientTransaction, error) {
	if err := n.UpdateEthTransactionsByAlchemy(req); err != nil {
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

func (n *NService) UpdateEthTransactionsByAlchemy(req request.GetEthTransactions) (err error) {
	client.URL = constant.GetAlchemyRPCUrlByNetwork(req.ChainId)

	var fromRpcAlchemyTxs response.RPCAlchemyTransactionDetails
	var toRpcAlchemyTxs response.RPCAlchemyTransactionDetails
	var transfers []response.RPCAlchemyTransactionTransfer
	fromPayload := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "alchemy_getAssetTransfers",
		"params": []map[string]interface{}{
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

	toPayload := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "alchemy_getAssetTransfers",
		"params": []map[string]interface{}{
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
		model, err := n.DecodeEthTransactionForAlchemy(req.ChainId, req.Address, v)
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			continue
		}
		saveTxs = append(saveTxs, model)
	}

	// get pending transaction
	pendingTxs, err := n.GetEthPendingTransactions(req)
	if err == nil {
		saveTxs = append(saveTxs, pendingTxs...)
	}

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

func (n *NService) DecodeEthTransactionForRpc(chainId uint, address string, tx *response.RPCTransaction) (model response.ClientTransaction, err error) {
	model.ChainId = chainId
	model.Address = address
	model.Hash = tx.Hash

	if tx.Input == "0x" {
		model.Asset = constant.ETH
		value, err := utils.HexStringToBigInt(tx.Value)
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			return model, err
		}

		model.Amount = utils.CalculateBalance(value, 18)
		model.Category = "eth"
	} else {
		isSupportContract, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, tx.To)
		if !isSupportContract {
			return model, errors.New("no support the contract")
		}

		model.Asset = contractName

		_, _, _, transactionValue, err := erc20.DecodeERC20TransactionInputData(chainId, tx.Hash, tx.Input)
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			return model, err
		}

		model.Amount = utils.CalculateBalance(transactionValue, decimals)
		model.Category = "erc20"
	}

	_, contractName, _, _ := chainUtils.GetContractInfo(chainId, tx.To)

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

	model.Status = "Pending"
	model.BlockTimestamp = 0

	return
}

func (n *NService) DecodeEthTransactionForAlchemy(chainId uint, address string, tx response.RPCAlchemyTransactionTransfer) (model response.ClientTransaction, err error) {
	model.ChainId = chainId
	model.Address = address
	model.Hash = tx.Hash
	if tx.Asset == constant.ETH {
		model.Asset = tx.Asset

		hexString := strings.TrimPrefix(tx.RawContract.Value, "0x")
		bigIntValue, _ := new(big.Int).SetString(hexString, 16)
		model.Amount = utils.CalculateBalance(bigIntValue, 18)

	} else if tx.Category == "erc20" {
		isSupportContract, contractName, contractAddress, decimals := chainUtils.GetContractInfo(chainId, tx.RawContract.Address)
		if !isSupportContract {
			return
		}
		model.Asset = contractName
		model.ContractAddress = contractAddress

		hexString := strings.TrimPrefix(tx.RawContract.Value, "0x")
		bigIntValue, _ := new(big.Int).SetString(hexString, 16)
		model.Amount = utils.CalculateBalance(bigIntValue, decimals)
	} else {
		global.NODE_LOG.Error(tx.Hash)
		return
	}

	_, contractName, _, _ := chainUtils.GetContractInfo(chainId, tx.To)

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
