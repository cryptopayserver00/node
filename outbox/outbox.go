package outbox

import (
	"context"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/model"
	"node/model/common"
	"strconv"
	"strings"
	"time"
)

const maxAttempts = 20

// EnqueueOwnTx 写入 WS / Webhook outbox 并尝试投递。
func EnqueueOwnTx(ctx context.Context, ownTxID uint) error {
	if ownTxID == 0 {
		return nil
	}
	payload := strconv.FormatUint(uint64(ownTxID), 10)

	var firstErr error
	if err := enqueueOne(ctx, ownTxID, payload, constant.WS_NOTIFICATION, ""); err != nil && firstErr == nil {
		firstErr = err
	}

	wh := global.NODE_CONFIG.Webhook
	if wh.Enabled {
		for _, raw := range wh.Urls {
			url := strings.TrimSpace(raw)
			if url == "" {
				continue
			}
			if err := enqueueOne(ctx, ownTxID, payload, constant.WEBHOOK_NOTIFICATION, url); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func enqueueOne(ctx context.Context, ownTxID uint, payload, channel, target string) error {
	// 避免同一 own_tx + channel + target 重复 pending/sent 堆积
	var count int64
	q := global.NODE_DB.WithContext(ctx).Model(&model.NotificationOutbox{}).
		Where("own_tx_id = ? AND channel = ? AND target = ? AND state IN ?", ownTxID, channel, target, []string{model.OutboxStatePending, model.OutboxStateSent})
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	row := model.NotificationOutbox{
		OwnTxID:    ownTxID,
		Payload:    payload,
		Channel:    channel,
		Target:     target,
		State:      model.OutboxStatePending,
		Attempts:   0,
		NODE_MODEL: common.NODE_MODEL{Status: 1},
	}
	if err := global.NODE_DB.WithContext(ctx).Create(&row).Error; err != nil {
		return fmt.Errorf("create outbox: %w", err)
	}
	return deliver(ctx, &row)
}

func deliver(ctx context.Context, row *model.NotificationOutbox) error {
	row.Attempts++
	var err error
	switch row.Channel {
	case constant.WEBHOOK_NOTIFICATION:
		err = deliverWebhook(ctx, row)
	default:
		err = deliverRedis(ctx, row)
	}
	if err != nil {
		_ = markRetry(ctx, row, err.Error())
		return err
	}
	now := time.Now()
	return global.NODE_DB.WithContext(ctx).Model(row).Updates(map[string]any{
		"state":         model.OutboxStateSent,
		"attempts":      row.Attempts,
		"last_error":    "",
		"sent_at":       now,
		"next_retry_at": nil,
	}).Error
}

func deliverRedis(ctx context.Context, row *model.NotificationOutbox) error {
	if global.NODE_REDIS == nil {
		return fmt.Errorf("redis not initialized")
	}
	return global.NODE_REDIS.RPush(ctx, row.Channel, row.Payload).Err()
}

func markRetry(ctx context.Context, row *model.NotificationOutbox, errMsg string) error {
	if len(errMsg) > 250 {
		errMsg = errMsg[:250]
	}
	state := model.OutboxStatePending
	if row.Attempts >= maxAttempts {
		state = model.OutboxStateFailed
	}
	backoff := time.Duration(row.Attempts) * 5 * time.Second
	if backoff > 5*time.Minute {
		backoff = 5 * time.Minute
	}
	next := time.Now().Add(backoff)
	return global.NODE_DB.WithContext(ctx).Model(row).Updates(map[string]any{
		"state":         state,
		"attempts":      row.Attempts,
		"last_error":    errMsg,
		"next_retry_at": next,
	}).Error
}

// RetryPending 重试到期的 pending outbox。
func RetryPending(ctx context.Context, limit int) {
	if global.NODE_DB == nil || limit <= 0 {
		return
	}
	var rows []model.NotificationOutbox
	now := time.Now()
	err := global.NODE_DB.WithContext(ctx).
		Where("state = ? AND (next_retry_at IS NULL OR next_retry_at <= ?) AND deleted_at IS NULL", model.OutboxStatePending, now).
		Order("id asc").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}
	for i := range rows {
		if err := deliver(ctx, &rows[i]); err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("outbox retry failed id=%d: %s", rows[i].ID, err.Error()))
		}
	}
}
