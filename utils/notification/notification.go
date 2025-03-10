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

func NotificationRequest(request request.NotificationRequest) (err error) {

	rd, _ := json.Marshal(request)
	global.NODE_LOG.Info("NotificationRequest: " + string(rd))

	err = handleNotification(request)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
	}

	utils.TxInformToTelegram("NotificationRequest: \n\n" + string(rd))

	return nil
}

func handleNotification(request request.NotificationRequest) (err error) {
	ownId, err := service.NodeService.SaveOwnTx(request)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	if ownId == 0 {
		global.NODE_LOG.Info(fmt.Sprintf("OwnId already existed, hash: %s", request.Hash))
		return
	}

	_, err = global.NODE_REDIS.RPush(context.Background(), constant.WS_NOTIFICATION, ownId).Result()
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	return nil
}
