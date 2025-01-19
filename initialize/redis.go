package initialize

import (
	"context"
	"node/global"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func Redis() {
	r := global.NODE_CONFIG.Redis
	client := redis.NewClient(&redis.Options{
		Addr:     r.Addr,
		Password: r.Password,
		DB:       r.DB,
	})

	if pong, err := client.Ping(context.Background()).Result(); err != nil {
		global.NODE_LOG.Error("redis connect failed: ", zap.Error(err))
	} else {
		global.NODE_LOG.Info("redis connect success: ", zap.String("pong", pong))
		global.NODE_REDIS = client
	}
}
