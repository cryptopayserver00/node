package core

import (
	"context"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/sweep/core/plugin"
	"node/sweep/setup"
	"node/utils"
	"node/utils/redisq"
	NODE_Client "node/utils/http"
	"strconv"
	"sync"
	"time"

)

func SetupBchLatestBlockHeight(ctx context.Context, client NODE_Client.Client, chainId uint) {
	var blockHeight int64
	switch global.NODE_CONFIG.BlockchainPlugin.BitcoinCash {
	case Tatum:
		// blockHeight = plugin.GetBchBlockHeightByTatum(ctx, client, chainId)
	case Mempool:
		blockHeight = plugin.GetBchBlockHeightByMempool(ctx, client, chainId)
	}

	if blockHeight > 0 {
		setup.SetupLatestBlockHeight(ctx, chainId, blockHeight)
		time.Sleep(10 * time.Second)
	}
}

func SweepBchBlockchainTransaction(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepBlockHeight, cacheBlockHeight *int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) {
	defer utils.HandlePanic()

	if len(*publicKey) <= 0 {
		SetupBchLatestBlockHeight(ctx, client, chainId)
		setup.UpdateCacheBlockHeight(ctx, chainId)
		setup.UpdateSweepBlockHeight(ctx, chainId)
		setup.UpdatePublicKey(ctx, chainId)
		return
	}

	if *sweepBlockHeight >= *cacheBlockHeight {
		SetupBchLatestBlockHeight(ctx, client, chainId)
		setup.UpdateCacheBlockHeight(ctx, chainId)
		setup.UpdatePublicKey(ctx, chainId)
		time.Sleep(time.Minute * 1)
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
			switch global.NODE_CONFIG.BlockchainPlugin.BitcoinCash {
			// case Tatum:
			// 	err = plugin.HandleBchBlockTransactionsByTatum(ctx, client, chainId, publicKey, height, constantPendingTransaction)
			case Mempool:
				err = plugin.HandleBchBlockTransactionsByMempool(ctx, client, chainId, publicKey, height, constantPendingTransaction)
			default:
				global.NODE_LOG.Error("not found the plugin of bitcoin cash chain")
				return
			}
			if err != nil {
				if _, rpushErr := redisq.EnqueueUnique(ctx, constantPendingBlock, strconv.FormatInt(height, 10)); rpushErr != nil {
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

func SweepBchBlockchainTransactionDetails(
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

	switch global.NODE_CONFIG.BlockchainPlugin.BitcoinCash {
	// case Tatum:
	// 	err = plugin.HandleBchTransactionDetailsByTatum(ctx, client, chainId, publicKey, constantPendingTransaction, txHash)
	case Mempool:
		err = plugin.HandleBchTransactionDetailsByMempool(ctx, client, chainId, publicKey, constantPendingTransaction, txHash)
	default:
		return
	}

	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("Can not handle the tx: %s, Retry | %s -> %s", txHash, constant.GetChainName(chainId), err.Error()))

		retryCount := (*txRetryCount)[txHash]
		retryCount++
		if retryCount >= setup.SweepThreshold {
			global.NODE_LOG.Error(fmt.Sprintf("give up tx after retries: %s", txHash))
			err = redisq.Ack(ctx, constantPendingTransaction, txHash)
			if err != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				time.Sleep(2 * time.Second)
				return
			}
			delete(*txRetryCount, txHash)
			return
		}
		(*txRetryCount)[txHash] = retryCount
		_ = redisq.Requeue(ctx, constantPendingTransaction, txHash)
		time.Sleep(2 * time.Second)
		return
	}

	err = redisq.Ack(ctx, constantPendingTransaction, txHash)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return
	}
	delete(*txRetryCount, txHash)
}

func SweepBchBlockchainPendingBlock(
	ctx context.Context,
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	retryBlockCount *map[int64]int,
	constantPendingBlock, constantPendingTransaction string,
) {
	defer utils.HandlePanic()

	blockHeight, err := redisq.Pop(ctx, constantPendingBlock)
	if err != nil {
		if redisq.IsEmpty(err) {
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
		_ = redisq.Ack(ctx, constantPendingBlock, blockHeight)
		return
	}

	switch global.NODE_CONFIG.BlockchainPlugin.BitcoinCash {
	// case Tatum:
	// 	err = plugin.HandleBchPendingBlockByTatum(ctx, client, chainId, publicKey, constantPendingBlock, constantPendingTransaction, blockHeight, blockHeightInt)
	case Mempool:
		err = plugin.HandleBchPendingBlockByMempool(ctx, client, chainId, publicKey, constantPendingBlock, constantPendingTransaction, blockHeight, blockHeightInt)
	default:
		return
	}

	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))

		retryCount := (*retryBlockCount)[blockHeightInt]
		retryCount++
		if retryCount >= setup.SweepThreshold {
			global.NODE_LOG.Error(fmt.Sprintf("give up block after retries: %d", blockHeightInt))
			err = redisq.Ack(ctx, constantPendingBlock, blockHeight)
			if err != nil {
				global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
				time.Sleep(2 * time.Second)
				return
			}
			delete(*retryBlockCount, blockHeightInt)
			return
		}
		(*retryBlockCount)[blockHeightInt] = retryCount
		_ = redisq.Requeue(ctx, constantPendingBlock, blockHeight)
		time.Sleep(2 * time.Second)
		return
	}

	err = redisq.Ack(ctx, constantPendingBlock, blockHeight)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		time.Sleep(2 * time.Second)
		return
	}
	delete(*retryBlockCount, blockHeightInt)
}
