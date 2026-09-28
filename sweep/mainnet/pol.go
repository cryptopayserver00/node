package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	polBlockRetryCount = make(map[int64]int)
	polTxRetryCount    = make(map[string]int)

	polClient NODE_Client.Client
)

func SweepPolBlockchain(ctx context.Context) {
	initPol(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepPolBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepPolBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepPolBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initPol(ctx context.Context) {
	core.SetupLatestBlockHeight(ctx, polClient, constant.POL_MAINNET)

	setup.SetupCacheBlockHeight(ctx, constant.POL_MAINNET)

	setup.SetupSweepBlockHeight(ctx, constant.POL_MAINNET)
}

func SweepPolBlockchainTransaction(ctx context.Context) {
	core.SweepBlockchainTransaction(
		ctx,
		polClient,
		constant.POL_MAINNET,
		&setup.PolPublicKey,
		&setup.PolSweepBlockHeight,
		&setup.PolCacheBlockHeight,
		constant.POL_SWEEP_BLOCK,
		constant.POL_PENDING_BLOCK,
		constant.POL_PENDING_TRANSACTION)
}

func SweepPolBlockchainTransactionDetails(ctx context.Context) {
	core.SweepBlockchainTransactionDetails(
		ctx,
		polClient,
		constant.POL_MAINNET,
		&setup.PolPublicKey,
		&polTxRetryCount,
		constant.POL_PENDING_TRANSACTION)
}

func SweepPolBlockchainPendingBlock(ctx context.Context) {
	core.SweepBlockchainPendingBlock(
		ctx,
		polClient,
		constant.POL_MAINNET,
		&setup.PolPublicKey,
		&polBlockRetryCount,
		constant.POL_PENDING_BLOCK,
		constant.POL_PENDING_TRANSACTION)
}
