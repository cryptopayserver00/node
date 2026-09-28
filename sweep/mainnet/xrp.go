package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	xrpBlockRetryCount = make(map[int64]int)
	xrpTxRetryCount    = make(map[string]int)

	xrpClient NODE_Client.Client
)

func SweepXrpBlockchain(ctx context.Context) {
	initXrp(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepXrpBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepXrpBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepXrpBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initXrp(ctx context.Context) {
	core.SetupXrpLatestBlockHeight(ctx, constant.XRP_MAINNET)

	setup.SetupCacheBlockHeight(ctx, constant.XRP_MAINNET)

	setup.SetupSweepBlockHeight(ctx, constant.XRP_MAINNET)
}

func SweepXrpBlockchainTransaction(ctx context.Context) {
	core.SweepXrpBlockchainTransaction(
		ctx,
		xrpClient,
		constant.XRP_MAINNET,
		&setup.XrpPublicKey,
		&setup.XrpSweepBlockHeight,
		&setup.XrpCacheBlockHeight,
		constant.XRP_SWEEP_BLOCK,
		constant.XRP_PENDING_BLOCK,
		constant.XRP_PENDING_TRANSACTION)
}

func SweepXrpBlockchainTransactionDetails(ctx context.Context) {
	core.SweepXrpBlockchainTransactionDetails(
		ctx,
		xrpClient,
		constant.XRP_MAINNET,
		&setup.XrpPublicKey,
		&xrpTxRetryCount,
		constant.XRP_PENDING_TRANSACTION)
}

func SweepXrpBlockchainPendingBlock(ctx context.Context) {
	core.SweepXrpBlockchainPendingBlock(
		ctx,
		xrpClient,
		constant.XRP_MAINNET,
		&setup.XrpPublicKey,
		&xrpBlockRetryCount,
		constant.XRP_PENDING_BLOCK,
		constant.XRP_PENDING_TRANSACTION)
}
