package model

import "node/model/common"

// SweepCursor 持久化每条链的扫块水位，Redis 仅作热缓存。
type SweepCursor struct {
	common.NODE_MODEL
	ChainId     uint   `json:"chain_id" gorm:"uniqueIndex:idx_sweep_cursor_chain;comment:chain_id"`
	NetworkName string `json:"network_name" gorm:"size:64;comment:network_name"`
	TipHeight   int64  `json:"tip_height" gorm:"comment:latest tip height"`
	CacheHeight int64  `json:"cache_height" gorm:"comment:cached tip for sweep lag"`
	SweepHeight int64  `json:"sweep_height" gorm:"comment:next height to sweep"`
}

func (SweepCursor) TableName() string {
	return "node_sweep_cursors"
}
