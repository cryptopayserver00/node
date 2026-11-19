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
	"node/model/node/response"
	"node/sweep/setup"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func (n *NService) BulkStorageUserWallets(c *gin.Context, wallets request.BulkStoreUserWallet) (resp response.BulkStoreUserWalletResponse, err error) {
	ctx := c.Request.Context()

	if len(wallets.BulkStorage) == 0 {
		return resp, nil
	}

	pendingRedis := make([]request.StoreUserWallet, 0, len(wallets.BulkStorage))

	err = global.NODE_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, v := range wallets.BulkStorage {
			created, saveErr := n.saveWalletDB(tx, v.ChainId, v.Address)
			if saveErr != nil {
				global.NODE_LOG.Error(fmt.Sprintf(
					"BulkStorageUserWallets saveWallet failed, chainId=%d, address=%s, err=%v",
					v.ChainId, v.Address, saveErr,
				))
				return fmt.Errorf("chain_id=%d address=%s: %w", v.ChainId, v.Address, saveErr)
			}
			if created {
				pendingRedis = append(pendingRedis, v)
			} else {
				// 已存在也确保 Redis 同步（提交后执行）
				pendingRedis = append(pendingRedis, v)
			}
		}
		return nil
	})

	if err != nil {
		return resp, errors.New("some wallets failed to store")
	}

	for _, v := range pendingRedis {
		if redisErr := setup.SavePublicKeyToRedis(ctx, v.ChainId, v.Address); redisErr != nil {
			global.NODE_LOG.Error(fmt.Sprintf(
				"BulkStorageUserWallets redis sync failed, chainId=%d, address=%s, err=%v",
				v.ChainId, v.Address, redisErr,
			))
			return resp, errors.New("wallets stored but redis sync failed")
		}
	}

	return resp, nil
}

func (n *NService) StoreUserWallet(c *gin.Context, wallet request.StoreUserWallet) (err error) {
	ctx := c.Request.Context()
	return n.saveWallet(ctx, global.NODE_DB.WithContext(ctx), wallet.ChainId, wallet.Address)
}

func (n *NService) HasWalletByChainIdAndAddress(ctx context.Context, chainId uint, address string) (hasWallet bool, err error) {
	return n.hasWallet(global.NODE_DB.WithContext(ctx), chainId, address)
}

func (n *NService) hasWallet(tx *gorm.DB, chainId uint, address string) (bool, error) {
	var count int64
	err := tx.Model(&model.Wallet{}).Where("chain_id = ? AND address = ?", chainId, address).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// saveWalletDB 只写库，返回是否新插入。
func (n *NService) saveWalletDB(tx *gorm.DB, chainId uint, address string) (created bool, err error) {
	if !constant.IsNetworkSupport(chainId) {
		return false, errors.New("do not support the network")
	}

	if !constant.IsAddressSupport(chainId, address) {
		return false, fmt.Errorf("do not support wallet address: id: %d, address: %s", chainId, address)
	}

	hasWallet, err := n.hasWallet(tx, chainId, address)
	if err != nil {
		return false, err
	}
	if hasWallet {
		return false, nil
	}

	saveWallet := model.Wallet{
		Address:     address,
		ChainId:     chainId,
		NetworkName: constant.GetChainName(chainId),
		NODE_MODEL:  common.NODE_MODEL{Status: 1},
	}
	if err = tx.Create(&saveWallet).Error; err != nil {
		if isDuplicateKeyErr(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (n *NService) saveWallet(ctx context.Context, tx *gorm.DB, chainId uint, address string) (err error) {
	_, err = n.saveWalletDB(tx, chainId, address)
	if err != nil {
		return err
	}
	return setup.SavePublicKeyToRedis(ctx, chainId, address)
}

func isDuplicateKeyErr(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique constraint")
}
