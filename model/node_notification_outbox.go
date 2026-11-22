package model

import (
	"time"

	"node/model/common"
)

// NotificationOutbox 可靠通知投递（先落库再推 Redis/Webhook，失败可重试）。
type NotificationOutbox struct {
	common.NODE_MODEL
	OwnTxID     uint       `json:"own_tx_id" gorm:"index;comment:own transaction id"`
	Payload     string     `json:"payload" gorm:"size:128;comment:queue payload, usually own_tx_id"`
	Channel     string     `json:"channel" gorm:"size:64;index;comment:WS_NOTIFICATION|WEBHOOK_NOTIFICATION"`
	Target      string     `json:"target" gorm:"size:512;comment:webhook url when channel=webhook"`
	State       string     `json:"state" gorm:"size:16;index;comment:pending|sent|failed"`
	Attempts    int        `json:"attempts" gorm:"comment:delivery attempts"`
	LastError   string     `json:"last_error" gorm:"size:255;comment:last error"`
	NextRetryAt *time.Time `json:"next_retry_at" gorm:"index;comment:next retry time"`
	SentAt      *time.Time `json:"sent_at" gorm:"comment:sent at"`
}

const (
	OutboxStatePending = "pending"
	OutboxStateSent    = "sent"
	OutboxStateFailed  = "failed"
)

func (NotificationOutbox) TableName() string {
	return "node_notification_outbox"
}
