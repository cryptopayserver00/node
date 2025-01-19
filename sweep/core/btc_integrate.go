package core

import (
	"context"
	"errors"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/sweep/core/plugin"
	"node/sweep/setup"
	"node/utils"
	NODE_Client "node/utils/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Tatum   = "tatum"
	Mempool = "mempool"
)

func SetupBtcLatestBlockHeight(client NODE_Client.Client, chainId uint) {

	var blockHeight int64
	switch global.NODE_CONFIG.BlockchainPlugin.Bitcoin {
	case Tatum:
		blockHeight = plugin.GetBtcBlockHeightByTatum(client, chainId)
	case Mempool:
		blockHeight = plugin.GetBtcBlockHeightByMempool(client, chainId)
	}

	if blockHeight > 0 {
		setup.SetupLatestBlockHeight(context.Background(), chainId, blockHeight)

		time.Sleep(10 * time.Second)
	}
}

func SweepBtcBlockchainTransaction(
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepCount *map[int64]int,
	sweepBlockHeight, cacheBlockHeight *int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) {
	defer utils.HandlePanic()

	if len(*publicKey) <= 0 {
		SetupBtcLatestBlockHeight(client, chainId)
		setup.UpdateCacheBlockHeight(context.Background(), chainId)
		setup.UpdateSweepBlockHeight(context.Background(), chainId)
		setup.UpdatePublicKey(context.Background(), chainId)
		return
	}

	if *sweepBlockHeight >= *cacheBlockHeight {
		SetupBtcLatestBlockHeight(client, chainId)
		setup.UpdateCacheBlockHeight(context.Background(), chainId)
		setup.UpdatePublicKey(context.Background(), chainId)
		time.Sleep(time.Minute * 1)
		return
	}

	var err error

	blockN, ok := (*sweepCount)[*sweepBlockHeight]
	if !ok {
		(*sweepCount)[*sweepBlockHeight] = 1
	} else if blockN >= setup.SweepThreshold {
		// skip current block
		_, err = global.NODE_REDIS.Set(context.Background(), constantSweepBlock, *sweepBlockHeight+1, 0).Result()
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return
		}

		// current block to pending queue
		_, err = global.NODE_REDIS.RPush(context.Background(), constantPendingBlock, *sweepBlockHeight).Result()
		if err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
			return
		}

		delete(*sweepCount, *sweepBlockHeight)

		*sweepBlockHeight += 1
		return
	} else {
		(*sweepCount)[*sweepBlockHeight]++
	}

	switch global.NODE_CONFIG.BlockchainPlugin.Bitcoin {
	case Tatum:
		plugin.HandleBtcBlockTransactionsByTatum(client, chainId, publicKey, sweepCount, sweepBlockHeight, constantSweepBlock, constantPendingTransaction)
		return
	case Mempool:
		plugin.HandleBtcBlockTransactionsByMempool(client, chainId, publicKey, sweepCount, sweepBlockHeight, constantSweepBlock, constantPendingTransaction)
		return
	}
}

func SweepBtcBlockchainTransactionDetails(
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	constantPendingTransaction string,
) {
	defer utils.HandlePanic()

	txHash, err := global.NODE_REDIS.LIndex(context.Background(), constantPendingTransaction, 0).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			time.Sleep(2 * time.Second)
			return
		}
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	switch global.NODE_CONFIG.BlockchainPlugin.Bitcoin {
	case Tatum:
		plugin.HandleBtcTransactionDetailsByTatum(client, chainId, publicKey, constantPendingTransaction, txHash)
		return
	case Mempool:
		plugin.HandleBtcTransactionDetailsByMempool(client, chainId, publicKey, constantPendingTransaction, txHash)
		return
	}
}

func SweepBtcBlockchainPendingBlock(
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	constantPendingBlock, constantPendingTransaction string,
) {
	defer utils.HandlePanic()

	blockHeight, err := global.NODE_REDIS.LIndex(context.Background(), constantPendingBlock, 0).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			time.Sleep(10 * time.Second)
			return
		}
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	blockHeightInt, err := strconv.ParseInt(blockHeight, 10, 64)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	switch global.NODE_CONFIG.BlockchainPlugin.Bitcoin {
	case Tatum:
		plugin.HandleBtcPendingBlockByTatum(client, chainId, publicKey, constantPendingBlock, constantPendingTransaction, blockHeight, blockHeightInt)
		return
	case Mempool:
		plugin.HandleBtcPendingBlockByMempool(client, chainId, publicKey, constantPendingBlock, constantPendingTransaction, blockHeight, blockHeightInt)
		return
	}
}
