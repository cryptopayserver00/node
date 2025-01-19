package main

import (
	"node/core"
	"node/global"
	"node/initialize"

	"go.uber.org/zap"
)

func main() {
	global.NODE_VP = core.Viper()
	global.NODE_LOG = core.Zap()
	zap.ReplaceGlobals(global.NODE_LOG)
	global.NODE_DB = initialize.Gorm()
	// initialize.Timer()
	// initialize.DBList()

	if global.NODE_DB != nil {
		initialize.RegisterTables()
		db, _ := global.NODE_DB.DB()
		defer db.Close()
	}

	core.RunWindowsServer()
}
