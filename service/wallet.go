package service

import (
	"context"
	"errors"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/model"
	"node/model/node/request"
	"node/model/node/response"
	"node/sweep/setup"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (n *NService) BulkStorageUserWallets(c *gin.Context, wallets request.BulkStoreUserWallet) (resp response.BulkStoreUserWalletResponse, err error) {
	ctx := c.Request.Context()

	if len(wallets.BulkStorage) == 0 {
		return resp, nil
	}

	err = global.NODE_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, v := range wallets.BulkStorage {
			if saveErr := n.saveWallet(ctx, tx, v.ChainId, v.Address); saveErr != nil {
				global.NODE_LOG.Error(fmt.Sprintf(
					"BulkStorageUserWallets saveWallet failed, chainId=%d, address=%s, err=%v",
					v.ChainId, v.Address, err,
				))
				return fmt.Errorf("chain_id=%d address=%s: %w", v.ChainId, v.Address, saveErr)
			}
		}
		return nil
	})

	if err != nil {
		return resp, errors.New("some wallets failed to store")
	}

	return resp, nil
}

func (n *NService) StoreUserWallet(c *gin.Context, wallet request.StoreUserWallet) (err error) {
	ctx := c.Request.Context()
	return n.saveWallet(ctx, global.NODE_DB.WithContext(ctx), wallet.ChainId, wallet.Address)
}

func (n *NService) HasWalletByChainIdAndAddress(ctx context.Context, chainId uint, address string) (hasWallet bool, err error) {
	var count int64

	err = global.NODE_DB.WithContext(ctx).Model(&model.Wallet{}).Where("chain_id = ? AND address = ?", chainId, address).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (n *NService) saveWallet(ctx context.Context, tx *gorm.DB, chainId uint, address string) (err error) {
	if !constant.IsNetworkSupport(chainId) {
		return errors.New("do not support the network")
	}

	if !constant.IsAddressSupport(chainId, address) {
		return fmt.Errorf("do not support wallet address: id: %d, address: %s", chainId, address)
	}

	hasWallet, err := n.HasWalletByChainIdAndAddress(ctx, chainId, address)
	if err != nil {
		return
	}

	if hasWallet {
		return nil
	}

	var saveWallet model.Wallet
	saveWallet.Address = address
	saveWallet.ChainId = chainId
	saveWallet.NetworkName = constant.GetChainName(chainId)
	saveWallet.Status = 1

	if err = tx.Create(&saveWallet).Error; err != nil {
		return
	}

	return setup.SavePublicKeyToRedis(ctx, chainId, address)
}
