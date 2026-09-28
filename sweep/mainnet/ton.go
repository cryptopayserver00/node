package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	tonBlockRetryCount = make(map[int64]int)
	tonTxRetryCount    = make(map[string]int)

	tonClient NODE_Client.Client
)

func SweepTonBlockchain(ctx context.Context) {

	initTon(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepTonBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepTonBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepTonBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initTon(ctx context.Context) {
	core.SetupTonLatestBlockHeight(ctx, tonClient, constant.TON_MAINNET)

	setup.SetupCacheBlockHeight(ctx, constant.TON_MAINNET)

	setup.SetupSweepBlockHeight(ctx, constant.TON_MAINNET)
}

func SweepTonBlockchainTransaction(ctx context.Context) {
	core.SweepTonBlockchainTransaction(
		ctx,
		tonClient,
		constant.TON_MAINNET,
		&setup.TonPublicKey,
		&setup.TonSweepBlockHeight,
		&setup.TonCacheBlockHeight,
		constant.TON_SWEEP_BLOCK,
		constant.TON_PENDING_BLOCK,
		constant.TON_PENDING_TRANSACTION)
}

func SweepTonBlockchainTransactionDetails(ctx context.Context) {
	core.SweepTonBlockchainTransactionDetails(
		ctx,
		tonClient,
		constant.TON_MAINNET,
		&setup.TonPublicKey,
		&tonTxRetryCount,
		constant.TON_PENDING_TRANSACTION)
}

func SweepTonBlockchainPendingBlock(ctx context.Context) {
	core.SweepTonBlockchainPendingBlock(
		ctx,
		tonClient,
		constant.TON_MAINNET,
		&setup.TonPublicKey,
		&tonBlockRetryCount,
		constant.TON_PENDING_BLOCK,
		constant.TON_PENDING_TRANSACTION)
}
