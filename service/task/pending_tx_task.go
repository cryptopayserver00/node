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

func RunGetPendingTxNumberTask() {
	for {
		now := time.Now()

		nextHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location())
		durationUntilNextHour := nextHour.Sub(now)

		ticker := time.NewTicker(durationUntilNextHour)

		<-ticker.C

		RunPendingTxNumberCore()
	}
}

func RunPendingTxNumberCore() {
	global.NODE_LOG.Info("---------- Run Get Pending Transaction Number Task ----------")

	// pending transaction and block

	allPendingTxString := []string{}
	for _, v := range constant.AllPendingTx {
		len, err := global.NODE_REDIS.LLen(context.Background(), v).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			global.NODE_LOG.Error(err.Error())
			continue
		}

		allPendingTxString = append(allPendingTxString, fmt.Sprintf("%s: %d\n", v, len))
	}

	allPendingBlockString := []string{}
	for _, v := range constant.AllPendingBlock {
		len, err := global.NODE_REDIS.LLen(context.Background(), v).Result()
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

	utils.InformToTelegram(strings.Join(allString, ""))
}
