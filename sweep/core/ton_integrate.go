package core

import (
	"context"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/sweep/setup"
	"node/utils"
	"node/utils/redisq"
	NODE_Client "node/utils/http"
	"time"

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

	txHash, err := redisq.Pop(ctx, constantPendingTransaction)
	if err != nil {
		if redisq.IsEmpty(err) {
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

	// blockHeight, err := redisq.Pop(ctx, constantPendingBlock)
	// if err != nil {
	// 	if redisq.IsEmpty(err) {
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
