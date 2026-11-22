package model

import "node/model/common"

type Wallet struct {
	common.NODE_MODEL
	Address         string `json:"address" gorm:"uniqueIndex:idx_wallet_chain_addr,priority:2;comment:address"`
	ChainId         uint   `json:"chain_id" gorm:"uniqueIndex:idx_wallet_chain_addr,priority:1;comment:chain_id"`
	NetworkName     string `json:"network_name" gorm:"comment:network_name"`
	AccountRef      string `json:"account_ref" gorm:"size:64;index;comment:upstream account reference"`
	WatchFromHeight int64  `json:"watch_from_height" gorm:"comment:start watching from this height; 0=from current"`
}

func (Wallet) TableName() string {
	return "node_wallets"
}
