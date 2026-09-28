package service

import (
	"context"
	"errors"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/model"
	"node/model/common"
	"node/model/node/request"
	"strings"

	"github.com/go-sql-driver/mysql"
)

func (n *NService) SaveOwnTx(ctx context.Context, req request.NotificationRequest) (id uint, err error) {

	if !constant.IsNetworkSupport(req.Chain) {
		return 0, errors.New("do not support network")
	}

	exists, err := n.HasOwnTxByNotificationObj(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("check own tx exists failed: %w", err)
	}
	if exists {
		return 0, nil
	}

	ownTx := model.OwnTransaction{
		ChainId:        req.Chain,
		Hash:           req.Hash,
		Address:        req.Address,
		FromAddress:    req.FromAddress,
		ToAddress:      req.ToAddress,
		Token:          req.Token,
		TransactType:   req.TransactType,
		Amount:         req.Amount,
		BlockTimestamp: req.BlockTimestamp,
		NODE_MODEL: common.NODE_MODEL{
			Status: 1,
		},
	}

	if err = global.NODE_DB.WithContext(ctx).Create(&ownTx).Error; err != nil {
		// 唯一索引冲突时，视为已存在（并发安全）
		if isDuplicateKeyErr(err) {
			return 0, nil
		}
		return 0, err
	}

	return 0, fmt.Errorf("create own tx failed, hash=%s: %w", req.Hash, err)
}

func (n *NService) HasOwnTxByNotificationObj(ctx context.Context, req request.NotificationRequest) (bool, error) {
	var count int64

	err := global.NODE_DB.WithContext(ctx).Model(&model.OwnTransaction{}).
		Where("chain_id = ? AND hash = ? AND address = ? AND from_address = ? AND to_address = ? AND token = ? AND transact_type = ? AND amount = ? AND block_timestamp = ?", req.Chain, req.Hash, req.Address, req.FromAddress, req.ToAddress, req.Token, req.TransactType, req.Amount, req.BlockTimestamp).
		Limit(1).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// 根据你用的数据库驱动调整
func isDuplicateKeyErr(err error) bool {
	if err == nil {
		return false
	}
	// MySQL
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}
	// PostgreSQL
	// var pgErr *pgconn.PgError
	// if errors.As(err, &pgErr) && pgErr.Code == "23505" {
	// 	return true
	// }
	return false
}

func (n *NService) GetOwnTxById(ctx context.Context, id string) (findOwnTx model.OwnTransaction, err error) {
	err = global.NODE_DB.WithContext(ctx).Where("id = ?", id).First(&findOwnTx).Error
	return
}

func (n *NService) GetTransactionsByChainAndAddress(ctx context.Context, req request.TransactionsByChainAndAddress) ([]model.OwnTransaction, int64, error) {
	var txs []model.OwnTransaction

	var total int64
	limit := req.PageSize
	offset := req.PageSize * (req.Page - 1)
	db := global.NODE_DB.WithContext(ctx).Model(&model.OwnTransaction{})

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
