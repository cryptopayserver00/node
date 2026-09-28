package core

import (
	"context"
	"errors"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/sweep/setup"
	"node/utils"
	NODE_Client "node/utils/http"
	"time"

	"github.com/redis/go-redis/v9"
)

func SetupTonLatestBlockHeight(ctx context.Context, client NODE_Client.Client, chainId uint) {
}

func SweepTonBlockchainTransaction(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepBlockHeight, cacheBlockHeight *int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) {
	defer utils.HandlePanic()

	if len(*publicKey) <= 0 {
		SetupLatestBlockHeight(ctx, client, chainId)
		setup.UpdateCacheBlockHeight(ctx, chainId)
		setup.UpdateSweepBlockHeight(ctx, chainId)
		setup.UpdatePublicKey(ctx, chainId)
		return
	}

	if *sweepBlockHeight >= *cacheBlockHeight {
		SetupLatestBlockHeight(ctx, client, chainId)
		setup.UpdateCacheBlockHeight(ctx, chainId)
		setup.UpdatePublicKey(ctx, chainId)
		time.Sleep(time.Second * 5)
		return
	}
}

func SweepTonBlockchainTransactionDetails(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	txRetryCount *map[string]int,
	constantPendingTransaction string,
) {
	defer utils.HandlePanic()

	txHash, err := global.NODE_REDIS.LIndex(ctx, constantPendingTransaction, 0).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			time.Sleep(2 * time.Second)
			return
		}
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return
	}

	global.NODE_LOG.Info(fmt.Sprintf("%s -> handle tx: %s", constant.GetChainName(chainId), txHash))

}

func SweepTonBlockchainPendingBlock(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	retryBlockCount *map[int64]int,
	constantPendingBlock, constantPendingTransaction string,
) {
	defer utils.HandlePanic()

	// blockHeight, err := global.NODE_REDIS.LIndex(ctx, constantPendingBlock, 0).Result()
	// if err != nil {
	// 	if errors.Is(err, redis.Nil) {
	// 		time.Sleep(2 * time.Second)
	// 		return
	// 	}
	// 	global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
	// 	time.Sleep(10 * time.Second)
	// 	return
	// }

	// blockHeightInt, err := strconv.ParseInt(blockHeight, 10, 64)
	// if err != nil {
	// 	global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
	// 	return
	// }
}
