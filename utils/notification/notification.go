package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/service"
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

	return nil
}

func handleNotification(request request.NotificationRequest) (err error) {
	// err = service.NodeService.SaveTx(request.Chain, request.Hash)
	// if err != nil {
	// 	global.NODE_LOG.Error(err.Error())
	// 	return
	// }

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
