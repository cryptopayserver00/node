package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/model"
	"node/model/common"
	"node/model/node/request"
	"strings"

	"gorm.io/gorm"
)

// 去重键：所有参与比较的字段拼接后哈希，作为锁名
func dedupeKey(req request.NotificationRequest) string {
	raw := fmt.Sprintf("%v|%s|%s|%s|%s|%s|%v|%v|%v",
		req.Chain, req.Hash, req.Address, req.FromAddress, req.ToAddress,
		req.Token, req.TransactType, req.Amount, req.BlockTimestamp)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:]) // 64 字符，符合 GET_LOCK 长度限制
}

func (n *NService) SaveOwnTx(ctx context.Context, req request.NotificationRequest) (id uint, created bool, err error) {
	if !constant.IsNetworkSupport(req.Chain) {
		return 0, false, errors.New("do not support network")
	}

	lockKey := dedupeKey(req)

	err = global.NODE_DB.WithContext(ctx).Connection(func(conn *gorm.DB) error {
		// 关键：转成会话，避免 Select/Where/Model 在多次调用间残留
		conn = conn.Session(&gorm.Session{})

		// GET_LOCK / RELEASE_LOCK 必须在同一连接
		var got int
		if err := conn.Raw("SELECT GET_LOCK(?,5)", lockKey).Scan(&got).Error; err != nil {
			return fmt.Errorf("get lock failed: %w", err)
		}
		if got != 1 {
			return errors.New("acquire lock timeout")
		}
		defer func() {
			// 即使 ctx 已取消也要释放锁，否则锁会一直挂在这个连接上
			_ = conn.Session(&gorm.Session{Context: context.WithoutCancel(ctx)}).
				Exec("SELECT RELEASE_LOCK(?)", lockKey).Error
		}()

		exists, err := n.hasOwnTx(conn, req)
		if err != nil {
			return fmt.Errorf("check own tx exists failed: %w", err)
		}
		if exists {
			return nil // 已存在，created=false
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
			NODE_MODEL:     common.NODE_MODEL{Status: 1},
		}
		if err := conn.Create(&ownTx).Error; err != nil {
			return fmt.Errorf("create own tx failed, hash=%s: %w", req.Hash, err)
		}
		id, created = ownTx.ID, true
		return nil
	})
	return id, created, err
}

// 存在性检查：用 conn 以便在锁所在的同一连接上执行
func (n *NService) hasOwnTx(conn *gorm.DB, req request.NotificationRequest) (bool, error) {
	var one int
	err := conn.Model(&model.OwnTransaction{}).
		Select("1").
		Where("chain_id = ? AND hash = ? AND address = ? AND from_address = ? AND to_address = ? AND token = ? AND transact_type = ? AND amount = ? AND block_timestamp = ?",
			req.Chain, req.Hash, req.Address, req.FromAddress, req.ToAddress, req.Token, req.TransactType, req.Amount, req.BlockTimestamp).
		Limit(1).
		Scan(&one).Error
	return one == 1, err
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
