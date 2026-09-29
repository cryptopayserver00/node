package task

import (
	"context"
	"errors"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/utils"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const GetPendingTxNumberInterval = 1 * time.Hour

func RunGetPendingTxNumberTask(ctx context.Context) {
	ticker := time.NewTicker(GetPendingTxNumberInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			RunPendingTxNumberCore(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func RunPendingTxNumberCore(ctx context.Context) {
	defer utils.HandlePanic()

	global.NODE_LOG.Info("---------- Run Get Pending Transaction Number Task ----------")

	// pending transaction and block

	allPendingTxString := []string{}
	for _, v := range constant.AllPendingTx {
		len, err := global.NODE_REDIS.LLen(ctx, v).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			global.NODE_LOG.Error(err.Error())
			continue
		}

		allPendingTxString = append(allPendingTxString, fmt.Sprintf("%s: %d\n", v, len))
	}

	allPendingBlockString := []string{}
	for _, v := range constant.AllPendingBlock {
		len, err := global.NODE_REDIS.LLen(ctx, v).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			global.NODE_LOG.Error(err.Error())
			continue
		}

		allPendingBlockString = append(allPendingBlockString, fmt.Sprintf("%s: %d\n", v, len))
	}

	allString := []string{}
	allString = append(allString, "---------- Run Pending Transaction Task ----------")
	allString = append(allString, "\n\n")
	allString = append(allString, allPendingTxString...)
	allString = append(allString, "\n")
	allString = append(allString, allPendingBlockString...)

	go utils.InformToTelegram(strings.Join(allString, ""))
}
