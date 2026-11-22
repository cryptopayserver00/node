package model

import (
	"encoding/json"
	"node/model/common"
	"time"
)

// OwnTransaction 是监听地址的资金变动事件（非「一笔链上 tx 一行」）。
// 唯一键：chain + hash + address + direction + token + transfer_key
type OwnTransaction struct {
	common.NODE_MODEL
	ChainId        uint   `json:"chain_id" gorm:"uniqueIndex:idx_transfer_event,priority:1;comment:chain_id"`
	Hash           string `json:"hash" gorm:"size:128;uniqueIndex:idx_transfer_event,priority:2;comment:tx hash"`
	Address        string `json:"address" gorm:"size:128;uniqueIndex:idx_transfer_event,priority:3;comment:monitored address"`
	TransactType   string `json:"transact_type" gorm:"size:16;uniqueIndex:idx_transfer_event,priority:4;comment:send|receive"`
	Token          string `json:"token" gorm:"size:64;uniqueIndex:idx_transfer_event,priority:5;comment:asset symbol"`
	TransferKey    string `json:"transfer_key" gorm:"size:64;uniqueIndex:idx_transfer_event,priority:6;comment:stable transfer identity"`
	TransferIndex  int    `json:"transfer_index" gorm:"comment:vout/log/ix index; 0 means derived"`
	FromAddress    string `json:"from_address" gorm:"size:128;comment:from_address"`
	ToAddress      string `json:"to_address" gorm:"size:128;comment:to_address"`
	Amount         string `json:"amount" gorm:"size:64;comment:amount"`
	BlockNumber    int64  `json:"block_number" gorm:"index;comment:block_number"`
	BlockTimestamp int    `json:"block_timestamp" gorm:"comment:block_timestamp"`
	Confirmations  int    `json:"confirmations" gorm:"comment:confirmations"`
	TxStatus       string `json:"tx_status" gorm:"size:16;index;comment:detected|confirmed|reorged"`
}

const (
	TxStatusDetected  = "detected"
	TxStatusConfirmed = "confirmed"
	TxStatusReorged   = "reorged"
)

func (OwnTransaction) TableName() string {
	return "node_own_transactions"
}

// MarshalJSON 对外暴露 id / created_at，便于 WS 与上游对账。
func (o OwnTransaction) MarshalJSON() ([]byte, error) {
	type Alias OwnTransaction
	return json.Marshal(&struct {
		ID        uint      `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		*Alias
	}{
		ID:        o.ID,
		CreatedAt: o.CreatedAt,
		Alias:     (*Alias)(&o),
	})
}
