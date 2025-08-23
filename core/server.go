package core

import (
	"fmt"
	"node/global"
	"node/initialize"
	"node/service"
	"node/service/task"
	"node/sweep"
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server interface {
	ListenAndServe() error
}

func RunWindowsServer() {
	if global.NODE_CONFIG.System.UseInit {
		if err := service.NodeService.InitChainList(); err != nil {
			global.NODE_LOG.Error(err.Error())
			return
		}
	}

	if global.NODE_CONFIG.System.UseRedis {
		initialize.Redis()
	}

	if global.NODE_CONFIG.System.UseMemcache {
		initialize.Memcache()
	}

	if global.NODE_CONFIG.Blockchain.OpenSweepBlock {
		sweep.RunBlockSweep()
	}

	if global.NODE_CONFIG.System.UseTask {
		task.RunTask()
	}

	router := initialize.Routers()

	address := fmt.Sprintf(":%d", global.NODE_CONFIG.System.Addr)
	server := initServer(address, router)

	global.NODE_LOG.Info("http server run success on ", zap.String("address", address))

	global.NODE_LOG.Error(server.ListenAndServe().Error())
}

func initServer(address string, router *gin.Engine) Server {
	server := endless.NewServer(address, router)
	server.ReadHeaderTimeout = 20 * time.Second
	server.WriteTimeout = 20 * time.Second
	server.MaxHeaderBytes = 1 << 20

	return server
}
