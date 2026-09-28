package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	arbitrumOneBlockRetryCount = make(map[int64]int)
	arbitrumOneTxRetryCount    = make(map[string]int)

	arbitrumOneClient NODE_Client.Client
)

func SweepArbitrumOneBlockchain(ctx context.Context) {

	initArbitrumOne(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepArbitrumOneBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepArbitrumOneBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepArbitrumNovaBlockchainTransaction(ctx)
			}
		}
	}()
}

func initArbitrumOne(ctx context.Context) {
	core.SetupLatestBlockHeight(ctx, arbitrumOneClient, constant.ARBITRUM_ONE)

	setup.SetupCacheBlockHeight(ctx, constant.ARBITRUM_ONE)

	setup.SetupSweepBlockHeight(ctx, constant.ARBITRUM_ONE)
}

func SweepArbitrumOneBlockchainTransaction(ctx context.Context) {
	core.SweepBlockchainTransaction(
		ctx,
		arbitrumOneClient,
		constant.ARBITRUM_ONE,
		&setup.ArbitrumOnePublicKey,
		&setup.ArbitrumOneSweepBlockHeight,
		&setup.ArbitrumOneCacheBlockHeight,
		constant.ARBITRUM_ONE_SWEEP_BLOCK,
		constant.ARBITRUM_ONE_PENDING_BLOCK,
		constant.ARBITRUM_ONE_PENDING_TRANSACTION)
}

func SweepArbitrumOneBlockchainTransactionDetails(ctx context.Context) {
	core.SweepBlockchainTransactionDetails(
		ctx,
		arbitrumOneClient,
		constant.ARBITRUM_ONE,
		&setup.ArbitrumOnePublicKey,
		&arbitrumOneTxRetryCount,
		constant.ARBITRUM_ONE_PENDING_TRANSACTION)
}

func SweepArbitrumOneBlockchainPendingBlock(ctx context.Context) {
	core.SweepBlockchainPendingBlock(
		ctx,
		arbitrumOneClient,
		constant.ARBITRUM_ONE,
		&setup.ArbitrumOnePublicKey,
		&arbitrumOneBlockRetryCount,
		constant.ARBITRUM_ONE_PENDING_BLOCK,
		constant.ARBITRUM_ONE_PENDING_TRANSACTION)
}
