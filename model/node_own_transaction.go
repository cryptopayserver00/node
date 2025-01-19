package model

import "node/model/common"

type OwnTransaction struct {
	common.NODE_MODEL
	ChainId        uint   `json:"chain_id" gorm:"comment:chain_id"`
	Hash           string `json:"hash" gorm:"comment:hash"`
	Address        string `json:"address" gorm:"comment:address"`
	FromAddress    string `json:"from_address" gorm:"comment:from_address"`
	ToAddress      string `json:"to_address" gorm:"comment:to_address"`
	Token          string `json:"token" gorm:"comment:token"`
	TransactType   string `json:"transact_type" gorm:"comment:transact_type"`
	Amount         string `json:"amount" gorm:"comment:amount"`
	BlockTimestamp int    `json:"block_timestamp" gorm:"comment:block_timestamp"`
}

func (OwnTransaction) TableName() string {
	return "node_own_transactions"
}
