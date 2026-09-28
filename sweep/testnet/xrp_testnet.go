package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	xrpTestnetBlockRetryCount = make(map[int64]int)
	xrpTestnetTxRetryCount    = make(map[string]int)

	xrpTestnetClient NODE_Client.Client
)

func SweepXrpTestnetBlockchain(ctx context.Context) {
	initXrpTestnet(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepXrpTestnetBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepXrpTestnetBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepXrpTestnetBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initXrpTestnet(ctx context.Context) {
	core.SetupXrpLatestBlockHeight(ctx, constant.XRP_TESTNET)

	setup.SetupCacheBlockHeight(ctx, constant.XRP_TESTNET)

	setup.SetupSweepBlockHeight(ctx, constant.XRP_TESTNET)
}

func SweepXrpTestnetBlockchainTransaction(ctx context.Context) {
	core.SweepXrpBlockchainTransaction(
		ctx,
		xrpTestnetClient,
		constant.XRP_TESTNET,
		&setup.XrpTestnetPublicKey,
		&setup.XrpTestnetSweepBlockHeight,
		&setup.XrpTestnetCacheBlockHeight,
		constant.XRP_TESTNET_SWEEP_BLOCK,
		constant.XRP_TESTNET_PENDING_BLOCK,
		constant.XRP_TESTNET_PENDING_TRANSACTION)
}

func SweepXrpTestnetBlockchainTransactionDetails(ctx context.Context) {
	core.SweepXrpBlockchainTransactionDetails(
		ctx,
		xrpTestnetClient,
		constant.XRP_TESTNET,
		&setup.XrpTestnetPublicKey,
		&xrpTestnetTxRetryCount,
		constant.XRP_TESTNET_PENDING_TRANSACTION)
}

func SweepXrpTestnetBlockchainPendingBlock(ctx context.Context) {
	core.SweepXrpBlockchainPendingBlock(
		ctx,
		xrpTestnetClient,
		constant.XRP_TESTNET,
		&setup.XrpTestnetPublicKey,
		&xrpTestnetBlockRetryCount,
		constant.XRP_TESTNET_PENDING_BLOCK,
		constant.XRP_TESTNET_PENDING_TRANSACTION)
}
