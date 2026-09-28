package core

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"node/global"
	"node/initialize"
	"node/service"
	"node/service/task"
	"node/sweep"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

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
		sweep.RunBlockSweep(context.Background())
	}

	if global.NODE_CONFIG.System.UseTask {
		task.RunTask(context.Background())
	}

	router := initialize.Routers()

	address := fmt.Sprintf(":%d", global.NODE_CONFIG.System.Addr)
	server := initServer(address, router)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		global.NODE_LOG.Info("server run success", zap.String("address", address))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			global.NODE_LOG.Error(err.Error())
		}
	}()

	<-ctx.Done()
	stop()
	global.NODE_LOG.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		global.NODE_LOG.Error("server forced to shutdown", zap.String("err", err.Error()))
	}

	global.NODE_LOG.Info("server exited")
}

func initServer(address string, router *gin.Engine) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           router,
		ReadHeaderTimeout: 20 * time.Second,
		WriteTimeout:      20 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
}
