package api

import (
	"context"
	"encoding/json"
	"net/http"
	"node/global"
	"node/global/constant"
	"node/service"
	"node/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	pingPeriod = 25 * time.Second // 必须小于代理的空闲超时
	pongWait   = 60 * time.Second
	writeWait  = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	// ReadBufferSize:  0,
	// WriteBufferSize: 0,
}

func (n *NodeApi) WsForTxInfo(c *gin.Context) {
	defer utils.HandlePanic()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	c.Request.Header.Add("Connection", "upgrade")
	c.Request.Header.Add("Upgrade", "websocket")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}
	defer conn.Close()

	// 读超时 + Pong 续期
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	// 读协程：只读，不写（客户端 Pong/关闭帧由这里处理）
	go func() {
		defer utils.HandlePanic()
		defer cancel()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	// 唯一的写协程（当前 goroutine）
	pingTicker := time.NewTicker(pingPeriod)
	taskTicker := time.NewTicker(500 * time.Millisecond)
	defer pingTicker.Stop()
	defer taskTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pingTicker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-taskTicker.C:
			if err := crycleTask(ctx, conn); err != nil {
				return // 写失败说明连接已坏
			}
		}
	}

	// done := make(chan struct{})

	// go func(conn *websocket.Conn) {
	// 	defer utils.HandlePanic()
	// 	defer close(done)

	// 	for {
	// 		messageType, message, err := conn.ReadMessage()
	// 		if err != nil {
	// 			return
	// 		}

	// 		err = conn.WriteMessage(messageType, message)
	// 		if err != nil {
	// 			global.NODE_LOG.Error(err.Error())
	// 			return
	// 		}
	// 	}
	// }(conn)

	// for {
	// 	select {
	// 	case <-done:
	// 		return
	// 	default:
	// 		crycleTask(ctx, conn)
	// 	}
	// }
}

func crycleTask(ctx context.Context, conn *websocket.Conn) error {
	// global.NODE_MUTEX.Lock()
	// defer global.NODE_MUTEX.Unlock()

	id, err := global.NODE_REDIS.LIndex(ctx, constant.WS_NOTIFICATION, 0).Result()
	if err != nil {
		return nil // 队列为空(redis.Nil)等，不算连接错误
	}

	tx, err := service.NodeService.GetOwnTxById(ctx, id)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return nil
	}

	data, err := json.Marshal(tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return nil
	}

	conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		global.NODE_LOG.Error(err.Error())
		return err
	}

	if _, err = global.NODE_REDIS.LPop(ctx, constant.WS_NOTIFICATION).Result(); err != nil {
		global.NODE_LOG.Error(err.Error())
	}
	return nil
}
