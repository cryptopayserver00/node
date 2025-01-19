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
	"node/utils"
	"strconv"
	"strings"

	"github.com/bradfitz/gomemcache/memcache"
)

var (
	// tx-chainId-coin(bnb,bep20)-address
	constantBscTransactionHistory = "txs-%d-%s-%s"
	bnb                           = "bnb"
	bep20                         = "bep20"
	BscTokenAction                = "tokentx"
	BscBnbAction                  = "txlist"
)

func (n *NService) GetBscTransactions(req request.GetBscTransactions) ([]response.ClientTransaction, error) {
	var action string

	if req.Action == BscTokenAction && req.ContractAddress != "" {
		action = bep20
	} else if req.Action == BscBnbAction {
		action = bnb
	} else {
		return nil, errors.New("no support the action or contractAddress")
	}

	if err := n.UpdateBscTransactions(req, action); err != nil {
		global.NODE_LOG.Error(err.Error())
	}

	item, err := global.NODE_MEMCACHE.Get(fmt.Sprintf(constantBscTransactionHistory, req.ChainId, action, req.Address))
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

	if req.Action == BscTokenAction {
		for _, v := range txs {
			if strings.EqualFold(v.ContractAddress, req.ContractAddress) {
				filterTxs = append(filterTxs, v)
			}
		}
		return filterTxs, nil
	} else {
		return txs, nil
	}
}

func (n *NService) UpdateBscTransactions(req request.GetBscTransactions, action string) (err error) {
	var (
		saveTxs []response.ClientTransaction
		page    = 1
		offset  = 300
		sort    = "desc"
	)

	client.URL = fmt.Sprintf("%s/api?module=account&action=%s&address=%s&startblock=0x0&endblock=latest&page=%d&offset=%d&sort=%s", constant.GetBscscanUrlByNetwork(req.ChainId), req.Action, req.Address, page, offset, sort)
	var transactionResponse response.BscscanTransactionResponse
	err = client.HTTPGet(&transactionResponse)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	if len(transactionResponse.Result) == 0 || transactionResponse.Status != "1" {
		return
	}

	for _, v := range transactionResponse.Result {
		tx, decodeErr := n.DecodeBscTransaction(req.ChainId, req.Address, req.Action, v)
		if decodeErr != nil {
			global.NODE_LOG.Error(decodeErr.Error())
			continue
		}
		saveTxs = append(saveTxs, tx)
	}

	byteTxs, err := json.Marshal(saveTxs)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	err = global.NODE_MEMCACHE.Set(&memcache.Item{
		Key:        fmt.Sprintf(constantBscTransactionHistory, req.ChainId, action, req.Address),
		Value:      byteTxs,
		Expiration: 86400,
	})
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	return nil
}

func (n *NService) DecodeBscTransaction(chainId uint, address, action string, tx response.BscscanTransactionResult) (model response.ClientTransaction, err error) {
	model.ChainId = chainId
	model.Address = address
	model.Hash = tx.Hash

	var contractAddress string

	if action == BscBnbAction {
		if tx.Input != "0x" {
			return model, errors.New("no support the tx")
		}
		contractAddress = "0x0000000000000000000000000000000000000000"
		if tx.TxreceiptStatus == "1" {
			model.Status = "Success"
		} else {
			model.Status = "Failed"
		}
		// isSupportContract, contractName, _, decimals := chainUtils.GetContractInfo(chainId, "0x0000000000000000000000000000000000000000")
		// if !isSupportContract {
		// 	return
		// }
		// model.Asset = contractName
		// bigIntValue, _ := new(big.Int).SetString(tx.Value, 10)
		// model.Amount = utils.CalculateBalance(bigIntValue, decimals)

		// if strings.EqualFold(tx.From, address) {
		// 	model.Type = "Send"
		// } else if strings.EqualFold(tx.To, address) {
		// 	model.Type = "Received"
		// } else {
		// 	return
		// }
	} else if action == BscTokenAction {
		contractAddress = tx.ContractAddress
		model.Status = "Success"
	} else {
		return model, errors.New("no support the tx")
	}

	isSupportContract, contractName, _, decimals := chainUtils.GetContractInfo(chainId, contractAddress)
	if !isSupportContract {
		return model, errors.New("no support the tx")
	}
	model.Asset = contractName
	bigIntValue, _ := new(big.Int).SetString(tx.Value, 10)
	model.Amount = utils.CalculateBalance(bigIntValue, decimals)
	model.ContractAddress = contractAddress
	if strings.EqualFold(tx.From, address) {
		model.Type = "Send"
	} else if strings.EqualFold(tx.To, address) {
		model.Type = "Received"
	} else {
		return model, errors.New("no support the tx")
	}

	timeNum, _ := strconv.Atoi(tx.TimeStamp)
	if timeNum != 0 {
		model.BlockTimestamp = timeNum * 1000
	} else {
		model.BlockTimestamp = 0
	}

	return
}
