package service

import (
	"context"
	"errors"
	"node/global"
	"node/model"
	"node/sweep/setup"
)

type SweepCursorStatus struct {
	ChainId     uint   `json:"chain_id"`
	NetworkName string `json:"network_name"`
	TipHeight   int64  `json:"tip_height"`
	CacheHeight int64  `json:"cache_height"`
	SweepHeight int64  `json:"sweep_height"`
	Lag         int64  `json:"lag"`
}

type OutboxStats struct {
	Pending int64 `json:"pending"`
	Sent    int64 `json:"sent"`
	Failed  int64 `json:"failed"`
}

type OpsStatus struct {
	Cursors    []SweepCursorStatus        `json:"cursors"`
	Outbox     OutboxStats                `json:"outbox"`
	FailedRows []model.NotificationOutbox `json:"failed_rows,omitempty"`
}

func (n *NService) GetOpsStatus(ctx context.Context, includeFailed bool) (OpsStatus, error) {
	var status OpsStatus

	var cursors []model.SweepCursor
	if err := global.NODE_DB.WithContext(ctx).Order("chain_id asc").Find(&cursors).Error; err != nil {
		return status, err
	}
	status.Cursors = make([]SweepCursorStatus, 0, len(cursors))
	for _, c := range cursors {
		tip, cache, sweep := c.TipHeight, c.CacheHeight, c.SweepHeight
		if st, ok := setup.GetChainState(c.ChainId); ok {
			if *st.Latest > 0 {
				tip = *st.Latest
			}
			if *st.Cache > 0 {
				cache = *st.Cache
			}
			if *st.Sweep > 0 {
				sweep = *st.Sweep
			}
		}
		lag := tip - sweep
		if lag < 0 {
			lag = 0
		}
		status.Cursors = append(status.Cursors, SweepCursorStatus{
			ChainId:     c.ChainId,
			NetworkName: c.NetworkName,
			TipHeight:   tip,
			CacheHeight: cache,
			SweepHeight: sweep,
			Lag:         lag,
		})
	}

	_ = global.NODE_DB.WithContext(ctx).Model(&model.NotificationOutbox{}).
		Where("state = ?", model.OutboxStatePending).Count(&status.Outbox.Pending).Error
	_ = global.NODE_DB.WithContext(ctx).Model(&model.NotificationOutbox{}).
		Where("state = ?", model.OutboxStateSent).Count(&status.Outbox.Sent).Error
	_ = global.NODE_DB.WithContext(ctx).Model(&model.NotificationOutbox{}).
		Where("state = ?", model.OutboxStateFailed).Count(&status.Outbox.Failed).Error

	if includeFailed {
		_ = global.NODE_DB.WithContext(ctx).
			Where("state = ?", model.OutboxStateFailed).
			Order("id desc").
			Limit(50).
			Find(&status.FailedRows).Error
	}
	return status, nil
}

func (n *NService) RetryFailedOutbox(ctx context.Context, id uint) error {
	res := global.NODE_DB.WithContext(ctx).Model(&model.NotificationOutbox{}).
		Where("id = ? AND state = ?", id, model.OutboxStateFailed).
		Updates(map[string]any{
			"state":         model.OutboxStatePending,
			"next_retry_at": nil,
			"last_error":    "",
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("outbox not found or not failed")
	}
	return nil
}
