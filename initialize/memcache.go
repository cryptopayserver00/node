package initialize

import (
	"node/global"

	"github.com/bradfitz/gomemcache/memcache"
	"go.uber.org/zap"
)

func Memcache() {
	m := global.NODE_CONFIG.Memcache
	client := memcache.New(m.Addr)

	if err := client.Ping(); err != nil {
		global.NODE_LOG.Error("memcache connect failed: ", zap.Error(err))
	} else {
		global.NODE_LOG.Info("memcache connect success: pong")
		global.NODE_MEMCACHE = client
	}
}
