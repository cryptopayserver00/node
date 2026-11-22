package model

import "node/model/common"

// Asset 链上资产元数据（native / erc20 / trc20 / spl …）。
type Asset struct {
	common.NODE_MODEL
	ChainId  uint   `json:"chain_id" gorm:"uniqueIndex:idx_asset_chain_contract,priority:1;comment:chain_id"`
	Contract string `json:"contract" gorm:"size:128;uniqueIndex:idx_asset_chain_contract,priority:2;comment:contract or mint"`
	Symbol   string `json:"symbol" gorm:"size:32;index;comment:symbol"`
	Decimals int    `json:"decimals" gorm:"comment:decimals"`
	Standard string `json:"standard" gorm:"size:16;comment:native|erc20|trc20|spl|other"`
	IsNative bool   `json:"is_native" gorm:"comment:is native coin"`
	Enabled  bool   `json:"enabled" gorm:"index;comment:enabled for sweep matching"`
}

const (
	AssetStandardNative = "native"
	AssetStandardERC20  = "erc20"
	AssetStandardTRC20  = "trc20"
	AssetStandardSPL    = "spl"
	AssetStandardOther  = "other"
)

func (Asset) TableName() string {
	return "node_assets"
}
