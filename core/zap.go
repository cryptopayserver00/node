package core

import (
	"fmt"
	"node/core/internal"
	"node/global"
	"node/utils"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Zap() *zap.Logger {
	dir := global.NODE_CONFIG.Zap.Director
	if ok, _ := utils.PathExists(dir); !ok {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			panic(fmt.Sprintf("create log dir failed: %v", err))
		}
	}

	cores := internal.Zap.GetZapCores()
	logger := zap.New(zapcore.NewTee(cores...))

	if global.NODE_CONFIG.Zap.ShowLine {
		logger = logger.WithOptions(zap.AddCaller())
	}

	hook := func(entry zapcore.Entry) error {
		if entry.Level < zapcore.ErrorLevel {
			return nil
		}
		// 切断递归:通知telegram模块自身产生的日志不再通知
		if strings.HasPrefix(entry.Message, "telegram") {
			return nil
		}

		caller := entry.Caller.TrimmedPath() // 例如 service/notify.go:42
		// 去重 key:级别 + 位置 + 消息,不含时间/随机数
		key := entry.Level.String() + "|" + caller + "|" + entry.Message

		text := fmt.Sprintf("[%s] %s\n%s\n%s",
			entry.Time.UTC().Format("2006-01-02 15:04:05"),
			entry.Level.CapitalString(),
			caller,
			entry.Message,
		)
		utils.InformToTelegramWithKey(key, text)
		return nil
	}

	return logger.WithOptions(zap.Hooks(hook))
}
