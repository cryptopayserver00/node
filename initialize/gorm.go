package initialize

import (
	"node/global"
	"node/model"
	"os"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Gorm() *gorm.DB {
	switch global.NODE_CONFIG.System.DbType {
	case "mysql":
		return GormMysql()
	default:
		return GormMysql()
	}
}

func RegisterTables() {
	db := global.NODE_DB
	if err := db.AutoMigrate(
		model.Wallet{},
		// model.Transaction{},
		model.OwnTransaction{},
	); err != nil {
		global.NODE_LOG.Error("db: register table failed", zap.Error(err))
		os.Exit(0)
	}

	global.NODE_LOG.Info("db: register table success")
}
