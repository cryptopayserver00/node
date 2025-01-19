package internal

import (
	"node/global"
	"os"

	"go.uber.org/zap/zapcore"
)

type fileRotatelogs struct{}

var FileRotatelogs = new(fileRotatelogs)

func (r *fileRotatelogs) GetWriteSyncer(level string) zapcore.WriteSyncer {
	fileWriter := NewCutter(global.NODE_CONFIG.Zap.Director, level, WithCutterFormat("2006-01-02"))
	if global.NODE_CONFIG.Zap.LogInConsole {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter))
	}
	return zapcore.AddSync(fileWriter)
}
