package core

import (
	"fmt"
	"node/core/internal"
	"node/global"
	"node/utils"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Zap() *zap.Logger {
	if ok, _ := utils.PathExists(global.NODE_CONFIG.Zap.Director); !ok {
		os.Mkdir(global.NODE_CONFIG.Zap.Director, os.ModePerm)
	}

	cores := internal.Zap.GetZapCores()
	logger := zap.New(zapcore.NewTee(cores...))

	if global.NODE_CONFIG.Zap.ShowLine {
		logger = logger.WithOptions(zap.AddCaller())
	}

	globalHook := func(entry zapcore.Entry) error {
		switch entry.Level {
		case zapcore.ErrorLevel:
			msg := fmt.Sprintf("[%s]\n\n%s | %s\n\n %s", entry.Time.UTC().Format("2006-01-02 15:04:05"), entry.Level.CapitalString(), utils.GenerateStringRandomly("node_", 12), entry.Message)
			go utils.InformToTelegram(msg)
		}
		return nil
	}

	logger = logger.WithOptions(zap.Hooks(globalHook))

	return logger
}
