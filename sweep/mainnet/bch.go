package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	bchBlockRetryCount = make(map[int64]int)
	bchTxRetryCount    = make(map[string]int)

	bchClient NODE_Client.Client
)

func SweepBchBlockchain(ctx context.Context) {
	initBch(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBchBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBchBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBchBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initBch(ctx context.Context) {
	core.SetupBchLatestBlockHeight(ctx, bchClient, constant.BCH_MAINNET)

	setup.SetupCacheBlockHeight(ctx, constant.BCH_MAINNET)

	setup.SetupSweepBlockHeight(ctx, constant.BCH_MAINNET)
}

func SweepBchBlockchainTransaction(ctx context.Context) {
	core.SweepBchBlockchainTransaction(
		ctx,
		bchClient,
		constant.BCH_MAINNET,
		&setup.BchPublicKey,
		&setup.BchSweepBlockHeight,
		&setup.BchCacheBlockHeight,
		constant.BCH_SWEEP_BLOCK,
		constant.BCH_PENDING_BLOCK,
		constant.BCH_PENDING_TRANSACTION)
}

func SweepBchBlockchainTransactionDetails(ctx context.Context) {
	core.SweepBchBlockchainTransactionDetails(
		ctx,
		bchClient,
		constant.BCH_MAINNET,
		&setup.BchPublicKey,
		&bchTxRetryCount,
		constant.BCH_PENDING_TRANSACTION)
}

func SweepBchBlockchainPendingBlock(ctx context.Context) {
	core.SweepBchBlockchainPendingBlock(
		ctx,
		bchClient,
		constant.BCH_MAINNET,
		&setup.BchPublicKey,
		&bchBlockRetryCount,
		constant.BCH_PENDING_BLOCK,
		constant.BCH_PENDING_TRANSACTION)
}
