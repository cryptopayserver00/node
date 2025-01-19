package model

import "node/model/common"

type Info struct {
	common.NODE_MODEL
	Url         string `json:"url" gorm:"comment:url"`
	Key         string `json:"key" gorm:"comment:key"`
	IsRPC       bool   `json:"is_rpc" gorm:"comment:is_rpc"`
	Network     string `json:"network" gorm:"comment:network"` // testnet mainnet
	NetworkName string `json:"network_name" gorm:"comment:network_name"`
}

func (Info) TableName() string {
	return "node_infos"
}
