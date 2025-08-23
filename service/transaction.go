package service

import (
	"errors"
	"node/global"
	"node/global/constant"
	"node/model"
	"node/model/node/request"
	"strings"

	"gorm.io/gorm"
)

func (n *NService) SaveOwnTx(request request.NotificationRequest) (id uint, err error) {
	if !constant.IsNetworkSupport(request.Chain) {
		return 0, errors.New("do not support network")
	}

	hasWallet, err := n.HasOwnTxByNotificationObj(request)
	if err != nil {
		return
	}

	if hasWallet {
		return 0, nil
	}

	var ownTx model.OwnTransaction
	ownTx.ChainId = request.Chain
	ownTx.Hash = request.Hash
	ownTx.Address = request.Address
	ownTx.FromAddress = request.FromAddress
	ownTx.ToAddress = request.ToAddress
	ownTx.Token = request.Token
	ownTx.TransactType = request.TransactType
	ownTx.Amount = request.Amount
	ownTx.BlockTimestamp = request.BlockTimestamp
	ownTx.Status = 1

	if err = global.NODE_DB.Create(&ownTx).Error; err != nil {
		return 0, err
	}

	return ownTx.ID, nil
}

func (n *NService) HasOwnTxByNotificationObj(request request.NotificationRequest) (hasWallet bool, err error) {
	var findOwnTx model.OwnTransaction

	err = global.NODE_DB.Where("chain_id = ? AND hash = ? AND address = ? AND from_address = ? AND to_address = ? AND token = ? AND transact_type = ? AND amount = ? AND block_timestamp = ?",
		request.Chain, request.Hash, request.Address, request.FromAddress, request.ToAddress, request.Token, request.TransactType, request.Amount, request.BlockTimestamp).First(&findOwnTx).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	if findOwnTx.ID > 0 {
		return true, nil
	}

	return false, nil
}

func (n *NService) GetOwnTxById(id string) (findOwnTx model.OwnTransaction, err error) {
	err = global.NODE_DB.Where("id = ?", id).First(&findOwnTx).Error
	return
}

func (n *NService) GetTransactionsByChainAndAddress(req request.TransactionsByChainAndAddress) ([]model.OwnTransaction, int64, error) {
	var txs []model.OwnTransaction

	var total int64
	limit := req.PageSize
	offset := req.PageSize * (req.Page - 1)
	db := global.NODE_DB.Model(&model.OwnTransaction{})

	if req.ChainIds != "" {
		chainIds := strings.Split(req.ChainIds, ",")
		db.Where("chain_id IN (?)", chainIds)
	}

	if req.Addresses != "" {
		addresses := strings.Split(req.Addresses, ",")
		db.Where("from_address IN (?) OR to_address IN (?)", addresses, addresses)
	}

	if err := db.Count(&total).Order("created_at desc").Offset(offset).Limit(limit).Find(&txs).Error; err != nil {
		return nil, total, err
	}

	return txs, total, nil
}
