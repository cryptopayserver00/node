package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	baseBlockRetryCount = make(map[int64]int)
	baseTxRetryCount    = make(map[string]int)

	baseClient NODE_Client.Client
)

func SweepBaseBlockchain(ctx context.Context) {
	initBase(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBaseBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBaseBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBaseBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initBase(ctx context.Context) {
	core.SetupLatestBlockHeight(ctx, baseClient, constant.BASE_MAINNET)

	setup.SetupCacheBlockHeight(ctx, constant.BASE_MAINNET)

	setup.SetupSweepBlockHeight(ctx, constant.BASE_MAINNET)
}

func SweepBaseBlockchainTransaction(ctx context.Context) {

	core.SweepBlockchainTransaction(
		ctx,
		baseClient,
		constant.BASE_MAINNET,
		&setup.BasePublicKey,
		&setup.BaseSweepBlockHeight,
		&setup.BaseCacheBlockHeight,
		constant.BASE_SWEEP_BLOCK,
		constant.BASE_PENDING_BLOCK,
		constant.BASE_PENDING_TRANSACTION)
}

func SweepBaseBlockchainTransactionDetails(ctx context.Context) {

	core.SweepBlockchainTransactionDetails(
		ctx,
		baseClient,
		constant.BASE_MAINNET,
		&setup.BasePublicKey,
		&baseTxRetryCount,
		constant.BASE_PENDING_TRANSACTION)
}

func SweepBaseBlockchainPendingBlock(ctx context.Context) {
	core.SweepBlockchainPendingBlock(
		ctx,
		baseClient,
		constant.BASE_MAINNET,
		&setup.BasePublicKey,
		&baseBlockRetryCount,
		constant.BASE_PENDING_BLOCK,
		constant.BASE_PENDING_TRANSACTION)
}
