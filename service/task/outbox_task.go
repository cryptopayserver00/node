package task

import (
	"context"
	"node/outbox"
	"time"
)

// RunOutboxRetry 后台重试通知 outbox。
func RunOutboxRetry(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				outbox.RetryPending(ctx, 100)
			}
		}
	}()
}
