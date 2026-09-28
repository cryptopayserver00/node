package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/service"
	"node/utils"
	NODE_Client "node/utils/http"
)

var (
	client NODE_Client.Client
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
	ownId, err := service.NodeService.SaveOwnTx(ctx, req)
	if err != nil {
		return fmt.Errorf("SaveOwnTx failed, hash=%s: %w", req.Hash, err)
	}

	if ownId == 0 {
		global.NODE_LOG.Info(fmt.Sprintf("OwnId already existed, hash: %s", req.Hash))
		return nil
	}

	if err := global.NODE_REDIS.RPush(ctx, constant.WS_NOTIFICATION, ownId).Err(); err != nil {
		return fmt.Errorf("RPush WS_NOTIFICATION failed, ownId=%d: %w", ownId, err)
	}

	return nil
}
