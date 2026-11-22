package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"node/config"
	"node/global"
	"node/model/node/request"
	"node/outbox"
	"node/service"
	"node/utils"
)

func NotificationRequest(ctx context.Context, req request.NotificationRequest) error {
	rd, err := json.Marshal(req)
	if err != nil {
		global.NODE_LOG.Error("NotificationRequest marshal failed: " + err.Error())
	} else {
		global.NODE_LOG.Info("NotificationRequest: " + string(rd))
		go utils.TxInformToTelegram("NotificationRequest: \n\n" + string(rd))
	}

	if err := handleNotification(ctx, req); err != nil {
		global.NODE_LOG.Error("handleNotification failed: " + err.Error())
		return err
	}

	return nil
}

func handleNotification(ctx context.Context, req request.NotificationRequest) error {
	ownId, created, err := service.NodeService.SaveOwnTx(ctx, req)
	if err != nil {
		return fmt.Errorf("SaveOwnTx failed, hash=%s: %w", req.Hash, err)
	}

	if !created {
		if ownId == 0 {
			global.NODE_LOG.Info(fmt.Sprintf("own tx skipped (empty/zero amount), hash: %s", req.Hash))
		} else {
			global.NODE_LOG.Info(fmt.Sprintf("own tx already existed, skip, hash: %s", req.Hash))
		}
		return nil
	}

	if global.NODE_CONFIG.Blockchain.EffectiveNotifyMode() == config.NotifyModeConfirmed {
		global.NODE_LOG.Info(fmt.Sprintf("own tx saved, defer notify until confirmed, id=%d hash=%s", ownId, req.Hash))
		return nil
	}

	if err := outbox.EnqueueOwnTx(ctx, ownId); err != nil {
		return fmt.Errorf("enqueue notification failed, ownId=%d: %w", ownId, err)
	}

	return nil
}
