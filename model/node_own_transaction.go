package model

import "node/model/common"

type OwnTransaction struct {
	common.NODE_MODEL
	ChainId        uint   `json:"chain_id" gorm:"uniqueIndex:idx_own_tx_dedupe,priority:1;comment:chain_id"`
	Hash           string `json:"hash" gorm:"uniqueIndex:idx_own_tx_dedupe,priority:2;comment:hash"`
	Address        string `json:"address" gorm:"uniqueIndex:idx_own_tx_dedupe,priority:3;comment:address"`
	FromAddress    string `json:"from_address" gorm:"comment:from_address"`
	ToAddress      string `json:"to_address" gorm:"comment:to_address"`
	Token          string `json:"token" gorm:"uniqueIndex:idx_own_tx_dedupe,priority:5;comment:token"`
	TransactType   string `json:"transact_type" gorm:"uniqueIndex:idx_own_tx_dedupe,priority:4;comment:transact_type"`
	Amount         string `json:"amount" gorm:"comment:amount"`
	BlockTimestamp int    `json:"block_timestamp" gorm:"comment:block_timestamp"`
}

func (OwnTransaction) TableName() string {
	return "node_own_transactions"
}
