package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	bscBlockRetryCount = make(map[int64]int)
	bscTxRetryCount    = make(map[string]int)

	bscClient NODE_Client.Client
)

func SweepBscBlockchain(ctx context.Context) {
	initBsc(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBscBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBscBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBscBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initBsc(ctx context.Context) {
	core.SetupLatestBlockHeight(ctx, bscClient, constant.BSC_MAINNET)

	setup.SetupCacheBlockHeight(ctx, constant.BSC_MAINNET)

	setup.SetupSweepBlockHeight(ctx, constant.BSC_MAINNET)
}

func SweepBscBlockchainTransaction(ctx context.Context) {
	core.SweepBlockchainTransaction(
		ctx,
		bscClient,
		constant.BSC_MAINNET,
		&setup.BscPublicKey,
		&setup.BscSweepBlockHeight,
		&setup.BscCacheBlockHeight,
		constant.BSC_SWEEP_BLOCK,
		constant.BSC_PENDING_BLOCK,
		constant.BSC_PENDING_TRANSACTION)
}

func SweepBscBlockchainTransactionDetails(ctx context.Context) {
	core.SweepBlockchainTransactionDetails(
		ctx,
		bscClient,
		constant.BSC_MAINNET,
		&setup.BscPublicKey,
		&bscTxRetryCount,
		constant.BSC_PENDING_TRANSACTION)
}

func SweepBscBlockchainPendingBlock(ctx context.Context) {
	core.SweepBlockchainPendingBlock(
		ctx,
		bscClient,
		constant.BSC_MAINNET,
		&setup.BscPublicKey,
		&bscBlockRetryCount,
		constant.BSC_PENDING_BLOCK,
		constant.BSC_PENDING_TRANSACTION)
}
