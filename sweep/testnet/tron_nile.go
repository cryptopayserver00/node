package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	tronNileBlockRetryCount = make(map[int64]int)
	tronNileTxRetryCount    = make(map[string]int)

	tronNileClient NODE_Client.Client
)

func SweepTronNileBlockchain(ctx context.Context) {
	initTronNile(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepTronNileBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepTronNileBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepTronNileBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initTronNile(ctx context.Context) {
	core.SetupTronLatestBlockHeight(ctx, tronNileClient, constant.TRON_NILE)

	setup.SetupCacheBlockHeight(ctx, constant.TRON_NILE)

	setup.SetupSweepBlockHeight(ctx, constant.TRON_NILE)
}

func SweepTronNileBlockchainTransaction(ctx context.Context) {
	core.SweepTronBlockchainTransaction(
		ctx,
		tronNileClient,
		constant.TRON_NILE,
		&setup.TronNilePublicKey,
		&setup.TronNileSweepBlockHeight,
		&setup.TronNileCacheBlockHeight,
		constant.TRON_NILE_SWEEP_BLOCK,
		constant.TRON_NILE_PENDING_BLOCK,
		constant.TRON_NILE_PENDING_TRANSACTION)
}

func SweepTronNileBlockchainTransactionDetails(ctx context.Context) {
	core.SweepTronBlockchainTransactionDetails(
		ctx,
		tronNileClient,
		constant.TRON_NILE,
		&setup.TronNilePublicKey,
		&tronNileTxRetryCount,
		constant.TRON_NILE_PENDING_TRANSACTION)
}

func SweepTronNileBlockchainPendingBlock(ctx context.Context) {
	core.SweepTronBlockchainPendingBlock(
		ctx,
		tronNileClient,
		constant.TRON_NILE,
		&setup.TronNilePublicKey,
		&tronNileBlockRetryCount,
		constant.TRON_NILE_PENDING_BLOCK,
		constant.TRON_NILE_PENDING_TRANSACTION)
}
