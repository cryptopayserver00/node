package task

import (
	"context"
	"errors"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/utils"
	"time"

	"github.com/redis/go-redis/v9"
)

const DailyReportInterval = 24 * time.Hour

func RunDailyReportTask(ctx context.Context) {
	ticker := time.NewTicker(DailyReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			RunDailyReportCore(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func RunDailyReportCore(ctx context.Context) {
	defer utils.HandlePanic()

	global.NODE_LOG.Info("---------- Run Daily Report Task ----------")

	count, err := global.NODE_REDIS.Get(ctx, constant.DAILY_REPORT_ERROR).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			count = "0"
		} else {
			RunDailyReportCore(ctx)
		}
	}

	message := fmt.Sprintf("[Daily Report] %s\n\nNumber of failures today: %s", time.Now().UTC().Format("2006-01-02 15:04:05"), count)

	if utils.InformToTelegram(message) {
		global.NODE_REDIS.Del(ctx, constant.DAILY_REPORT_ERROR)
	}
}
