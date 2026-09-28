package task

import (
	"context"
	NODE_Client "node/utils/http"
	"time"
)

var (
	client NODE_Client.Client
)

func RunTask(ctx context.Context) {
	go RunApiKeyTestTask(ctx)
	// go RunDailyReportTask(ctx)
	// go RunGetPendingTxNumberTask(ctx)
}

func waitUntilNextInterval(ctx context.Context, intervalSec int) bool {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()+intervalSec, 0, now.Location())
	duration := next.Sub(now)

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}
