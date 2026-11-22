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

	if err := migrateTransferEventSchema(db); err != nil {
		global.NODE_LOG.Error("db: migrate transfer events failed", zap.Error(err))
		os.Exit(0)
	}

	if err := db.AutoMigrate(
		model.Wallet{},
		model.OwnTransaction{},
		model.SweepCursor{},
		model.Asset{},
		model.NotificationOutbox{},
	); err != nil {
		global.NODE_LOG.Error("db: register table failed", zap.Error(err))
		os.Exit(0)
	}

	global.NODE_LOG.Info("db: register table success")
}
