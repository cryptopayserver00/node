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

// BuildTransferKey 生成事件幂等键，与唯一索引一致。
func BuildTransferKey(req request.NotificationRequest) string {
	if req.HasTransferIndex {
		return fmt.Sprintf("idx:%d", req.TransferIndex)
	}
	raw := fmt.Sprintf("%s|%s|%s|%s",
		strings.ToLower(req.FromAddress),
		strings.ToLower(req.ToAddress),
		req.Amount,
		strings.ToLower(req.TransactType),
	)
	sum := sha256.Sum256([]byte(raw))
	return "fp:" + hex.EncodeToString(sum[:16])
}

func dedupeLockKey(req request.NotificationRequest, transferKey string) string {
	raw := fmt.Sprintf("%d|%s|%s|%s|%s|%s",
		req.Chain, req.Hash, req.Address, req.TransactType, req.Token, transferKey)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (n *NService) SaveOwnTx(ctx context.Context, req request.NotificationRequest) (id uint, created bool, err error) {
	if !constant.IsNetworkSupport(req.Chain) {
		return 0, false, errors.New("do not support network")
	}
	if req.Hash == "" || req.Address == "" || req.TransactType == "" {
		return 0, false, errors.New("invalid transfer event fields")
	}
	// 零金额不作为有效资金事件（保留显式非空 amount 的业务数据）
	if req.Amount == "" || req.Amount == "0" || req.Amount == "0.0" || req.Amount == "0.00" {
		return 0, false, nil
	}

	transferKey := BuildTransferKey(req)
	lockKey := dedupeLockKey(req, transferKey)

	err = global.NODE_DB.WithContext(ctx).Connection(func(conn *gorm.DB) error {
		conn = conn.Session(&gorm.Session{})

		var got int
		if err := conn.Raw("SELECT GET_LOCK(?,5)", lockKey).Scan(&got).Error; err != nil {
			return fmt.Errorf("get lock failed: %w", err)
		}
		if got != 1 {
			return errors.New("acquire lock timeout")
		}
		defer func() {
			_ = conn.Session(&gorm.Session{Context: context.WithoutCancel(ctx)}).
				Exec("SELECT RELEASE_LOCK(?)", lockKey).Error
		}()

		exists, existID, err := n.hasOwnTx(conn, req, transferKey)
		if err != nil {
			return fmt.Errorf("check own tx exists failed: %w", err)
		}
		if exists {
			id = existID
			return nil
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
			BlockNumber:    req.BlockNumber,
			TransferIndex:  req.TransferIndex,
			TransferKey:    transferKey,
			Confirmations:  0,
			TxStatus:       model.TxStatusDetected,
			NODE_MODEL:     common.NODE_MODEL{Status: 1},
		}
		if err := conn.Create(&ownTx).Error; err != nil {
			if isDuplicateKeyErr(err) {
				_, existID, _ = n.hasOwnTx(conn, req, transferKey)
				id = existID
				return nil
			}
			return fmt.Errorf("create own tx failed, hash=%s: %w", req.Hash, err)
		}
		id, created = ownTx.ID, true
		return nil
	})
	return id, created, err
}

func (n *NService) hasOwnTx(conn *gorm.DB, req request.NotificationRequest, transferKey string) (bool, uint, error) {
	var row model.OwnTransaction
	err := conn.Select("id").
		Where("chain_id = ? AND hash = ? AND address = ? AND transact_type = ? AND token = ? AND transfer_key = ?",
			req.Chain, req.Hash, req.Address, req.TransactType, req.Token, transferKey).
		Limit(1).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}
	return true, row.ID, nil
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
		db = db.Where("chain_id IN (?)", chainIds)
	}

	if req.Addresses != "" {
		addresses := strings.Split(req.Addresses, ",")
		db = db.Where("from_address IN (?) OR to_address IN (?) OR address IN (?)", addresses, addresses, addresses)
	}

	if err := db.Count(&total).Order("created_at desc").Offset(offset).Limit(limit).Find(&txs).Error; err != nil {
		return nil, total, err
	}

	return txs, total, nil
}
