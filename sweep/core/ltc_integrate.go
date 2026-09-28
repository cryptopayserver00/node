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
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

func SetupLtcLatestBlockHeight(ctx context.Context, client NODE_Client.Client, chainId uint) {
	var blockHeight int64
	switch global.NODE_CONFIG.BlockchainPlugin.Litecoin {
	case Tatum:
		blockHeight = plugin.GetLtcBlockHeightByTatum(ctx, client, chainId)
	case Mempool:
		blockHeight = plugin.GetLtcBlockHeightByMempool(ctx, client, chainId)
	}

	if blockHeight > 0 {
		setup.SetupLatestBlockHeight(ctx, chainId, blockHeight)
		time.Sleep(10 * time.Second)
	}
}

func SweepLtcBlockchainTransaction(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepBlockHeight, cacheBlockHeight *int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) {
	defer utils.HandlePanic()

	if len(*publicKey) <= 0 {
		SetupLtcLatestBlockHeight(ctx, client, chainId)
		setup.UpdateCacheBlockHeight(ctx, chainId)
		setup.UpdateSweepBlockHeight(ctx, chainId)
		setup.UpdatePublicKey(ctx, chainId)
		return
	}

	if *sweepBlockHeight >= *cacheBlockHeight {
		SetupLtcLatestBlockHeight(ctx, client, chainId)
		setup.UpdateCacheBlockHeight(ctx, chainId)
		setup.UpdatePublicKey(ctx, chainId)
		time.Sleep(10 * time.Second)
		return
	}

	var wg sync.WaitGroup

	numWorkers := calcNumWorkers(*sweepBlockHeight, *cacheBlockHeight, bitcoinMaxWorkers)
	if numWorkers == 0 {
		return
	}

	start := *sweepBlockHeight
	end := start + int64(numWorkers) - 1

	for h := start; h <= end; h++ {
		wg.Add(1)
		go func(height int64) {
			defer wg.Done()
			defer utils.HandlePanic()

			var err error
			switch global.NODE_CONFIG.BlockchainPlugin.Litecoin {
			case Tatum:
				err = plugin.HandleLtcBlockTransactionsByTatum(ctx, client, chainId, publicKey, height, constantPendingTransaction)
			case Mempool:
				err = plugin.HandleLtcBlockTransactionsByMempool(ctx, client, chainId, publicKey, height, constantPendingTransaction)
			default:
				global.NODE_LOG.Error("not found the plugin of litecoin chain")
				return
			}
			if err != nil {
				if _, rpushErr := global.NODE_REDIS.RPush(ctx, constantPendingBlock, height).Result(); rpushErr != nil {
					global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), rpushErr.Error()))
				}
			}
		}(h)
	}

	wg.Wait()
	*sweepBlockHeight = end + 1
	if _, err := global.NODE_REDIS.Set(ctx, constantSweepBlock, *sweepBlockHeight, 0).Result(); err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}
}

func SweepLtcBlockchainTransactionDetails(
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

	switch global.NODE_CONFIG.BlockchainPlugin.Litecoin {
	case Tatum:
		err = plugin.HandleLtcTransactionDetailsByTatum(ctx, client, chainId, publicKey, constantPendingTransaction, txHash)
	case Mempool:
		err = plugin.HandleLtcTransactionDetailsByMempool(ctx, client, chainId, publicKey, constantPendingTransaction, txHash)
	default:
		return
	}

	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("Can not handle the tx: %s, Retry | %s -> %s", txHash, constant.GetChainName(chainId), err.Error()))

		retryCount := (*txRetryCount)[txHash]
		retryCount++
		if retryCount >= setup.SweepThreshold {
			global.NODE_LOG.Error(fmt.Sprintf("give up tx after retries: %s", txHash))
			_, err = global.NODE_REDIS.LPop(ctx, constantPendingTransaction).Result()
			if err != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				time.Sleep(2 * time.Second)
				return
			}
			delete(*txRetryCount, txHash)
			return
		}
		(*txRetryCount)[txHash] = retryCount
		time.Sleep(2 * time.Second)
		return
	}

	_, err = global.NODE_REDIS.LPop(ctx, constantPendingTransaction).Result()
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return
	}
	delete(*txRetryCount, txHash)
}

func SweepLtcBlockchainPendingBlock(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	retryBlockCount *map[int64]int,
	constantPendingBlock, constantPendingTransaction string,
) {
	defer utils.HandlePanic()

	blockHeight, err := global.NODE_REDIS.LIndex(ctx, constantPendingBlock, 0).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			time.Sleep(2 * time.Second)
			return
		}
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(10 * time.Second)
		return
	}

	blockHeightInt, err := strconv.ParseInt(blockHeight, 10, 64)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return
	}

	switch global.NODE_CONFIG.BlockchainPlugin.Litecoin {
	case Tatum:
		err = plugin.HandleLtcPendingBlockByTatum(ctx, client, chainId, publicKey, constantPendingBlock, constantPendingTransaction, blockHeight, blockHeightInt)
	case Mempool:
		err = plugin.HandleLtcPendingBlockByMempool(ctx, client, chainId, publicKey, constantPendingBlock, constantPendingTransaction, blockHeight, blockHeightInt)
	default:
		return
	}

	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))

		retryCount := (*retryBlockCount)[blockHeightInt]
		retryCount++
		if retryCount >= setup.SweepThreshold {
			global.NODE_LOG.Error(fmt.Sprintf("give up block after retries: %d", blockHeightInt))
			_, err = global.NODE_REDIS.LPop(ctx, constantPendingBlock).Result()
			if err != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				time.Sleep(2 * time.Second)
				return
			}
			delete(*retryBlockCount, blockHeightInt)
			return
		}
		(*retryBlockCount)[blockHeightInt] = retryCount
		time.Sleep(2 * time.Second)
		return
	}

	_, err = global.NODE_REDIS.LPop(ctx, constantPendingBlock).Result()
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return
	}
	delete(*retryBlockCount, blockHeightInt)
}
