package initialize

import (
	"net/http"
	"node/global"
	"node/middleware"
	"node/router"

	"github.com/gin-gonic/gin"
	// swaggerFiles "github.com/swaggo/files"
	// ginSwagger "github.com/swaggo/gin-swagger"
)

func Routers() *gin.Engine {

	SetGinMode(global.NODE_CONFIG.System.Env)

	newRouter := gin.New()
	newRouter.Use(middleware.Cors())

	newRouter.MaxMultipartMemory = 1 << 20

	// newRouter.GET(global.NODE_CONFIG.System.RouterPrefix+"/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// global.NODE_LOG.Info("register swagger handler")

	MainRouter := new(router.MainRouter)

	newRouter.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "ok")
	})

	Group := newRouter.Group(global.NODE_CONFIG.System.RouterPrefix)
	{
		Group.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, "ok")
		})

		MainRouter.InitRouter(Group)
	}

	global.NODE_LOG.Info("router register success")
	return newRouter
}

func SetGinMode(mode string) {
	switch mode {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}
