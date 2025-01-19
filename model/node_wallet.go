package model

import "node/model/common"

type Wallet struct {
	common.NODE_MODEL
	Address     string `json:"address" gorm:"comment:address"`
	ChainId     uint   `json:"chain_id" gorm:"comment:chain_id"`
	NetworkName string `json:"network_name" gorm:"comment:network_name"`
}

func (Wallet) TableName() string {
	return "node_wallets"
}
