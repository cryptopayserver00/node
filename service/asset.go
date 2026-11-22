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
	sweepUtils "node/sweep/utils"
	"strings"

	"gorm.io/gorm"
)

func guessAssetStandard(chainId uint, isNative bool) string {
	if isNative {
		return model.AssetStandardNative
	}
	switch chainId {
	case constant.ETH_MAINNET, constant.ETH_SEPOLIA,
		constant.BSC_MAINNET, constant.BSC_TESTNET,
		constant.OP_MAINNET, constant.OP_SEPOLIA,
		constant.ARBITRUM_ONE, constant.ARBITRUM_NOVA, constant.ARBITRUM_SEPOLIA,
		constant.POL_MAINNET, constant.POL_TESTNET,
		constant.AVAX_MAINNET, constant.AVAX_TESTNET,
		constant.BASE_MAINNET, constant.BASE_SEPOLIA:
		return model.AssetStandardERC20
	case constant.TRON_MAINNET, constant.TRON_NILE:
		return model.AssetStandardTRC20
	case constant.SOL_MAINNET, constant.SOL_DEVNET:
		return model.AssetStandardSPL
	default:
		return model.AssetStandardOther
	}
}

// SyncAssetsFromChainList 将 ChainList.json 资产 upsert 到 node_assets。
func SyncAssetsFromChainList() error {
	if global.NODE_DB == nil {
		return fmt.Errorf("db not ready")
	}
	if len(model.ChainList) == 0 {
		return nil
	}

	for _, chain := range model.ChainList {
		for _, coin := range chain.Coins {
			contract := strings.TrimSpace(coin.Contract)
			asset := model.Asset{
				ChainId:  chain.ChainId,
				Contract: contract,
				Symbol:   coin.Symbol,
				Decimals: coin.Decimals,
				Standard: guessAssetStandard(chain.ChainId, coin.IsMainCoin),
				IsNative: coin.IsMainCoin,
				Enabled:  true,
				NODE_MODEL: common.NODE_MODEL{
					Status: 1,
				},
			}

			var existing model.Asset
			err := global.NODE_DB.Where("chain_id = ? AND contract = ?", chain.ChainId, contract).
				Take(&existing).Error
			if err != nil {
				if createErr := global.NODE_DB.Create(&asset).Error; createErr != nil {
					return fmt.Errorf("create asset chain=%d contract=%s: %w", chain.ChainId, contract, createErr)
				}
				continue
			}

			if err := global.NODE_DB.Model(&existing).Updates(map[string]any{
				"symbol":    asset.Symbol,
				"decimals":  asset.Decimals,
				"standard":  asset.Standard,
				"is_native": asset.IsNative,
				"enabled":   true,
				"status":    1,
			}).Error; err != nil {
				return fmt.Errorf("update asset id=%d: %w", existing.ID, err)
			}
		}
	}

	global.NODE_LOG.Info(fmt.Sprintf("assets synced from chainlist, chains=%d", len(model.ChainList)))
	return nil
}

func (n *NService) ListAssets(ctx context.Context, req request.ListAssetsRequest) ([]model.Asset, error) {
	db := global.NODE_DB.WithContext(ctx).Model(&model.Asset{})
	if req.ChainId > 0 {
		db = db.Where("chain_id = ?", req.ChainId)
	}
	var rows []model.Asset
	if err := db.Order("chain_id asc, symbol asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (n *NService) UpsertAsset(ctx context.Context, req request.UpsertAssetRequest) (model.Asset, error) {
	if !constant.IsNetworkSupport(req.ChainId) {
		return model.Asset{}, errors.New("do not support the network")
	}
	contract := strings.TrimSpace(req.Contract)
	standard := strings.TrimSpace(req.Standard)
	if standard == "" {
		standard = guessAssetStandard(req.ChainId, req.IsNative)
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	var existing model.Asset
	err := global.NODE_DB.WithContext(ctx).
		Where("chain_id = ? AND contract = ?", req.ChainId, contract).
		Take(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		asset := model.Asset{
			ChainId:    req.ChainId,
			Contract:   contract,
			Symbol:     req.Symbol,
			Decimals:   req.Decimals,
			Standard:   standard,
			IsNative:   req.IsNative,
			Enabled:    enabled,
			NODE_MODEL: common.NODE_MODEL{Status: 1},
		}
		if createErr := global.NODE_DB.WithContext(ctx).Create(&asset).Error; createErr != nil {
			return model.Asset{}, createErr
		}
		_ = sweepUtils.ReloadContractIndexFromAssets()
		return asset, nil
	}
	if err != nil {
		return model.Asset{}, err
	}

	if err := global.NODE_DB.WithContext(ctx).Model(&existing).Updates(map[string]any{
		"symbol":    req.Symbol,
		"decimals":  req.Decimals,
		"standard":  standard,
		"is_native": req.IsNative,
		"enabled":   enabled,
		"status":    1,
	}).Error; err != nil {
		return model.Asset{}, err
	}
	_ = sweepUtils.ReloadContractIndexFromAssets()
	existing.Symbol = req.Symbol
	existing.Decimals = req.Decimals
	existing.Standard = standard
	existing.IsNative = req.IsNative
	existing.Enabled = enabled
	return existing, nil
}

func (n *NService) SetAssetEnabled(ctx context.Context, req request.SetAssetEnabledRequest) error {
	contract := strings.TrimSpace(req.Contract)
	res := global.NODE_DB.WithContext(ctx).Model(&model.Asset{}).
		Where("chain_id = ? AND contract = ?", req.ChainId, contract).
		Update("enabled", req.Enabled)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("asset not found")
	}
	return sweepUtils.ReloadContractIndexFromAssets()
}
