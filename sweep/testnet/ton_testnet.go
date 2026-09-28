package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	tonTestnetBlockRetryCount = make(map[int64]int)
	tonTestnetTxRetryCount    = make(map[string]int)

	tonTestnetClient NODE_Client.Client
)

func SweepTonTestnetBlockchain(ctx context.Context) {
	initTonTestnet(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepTonTestnetBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepTonTestnetBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepTonTestnetBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initTonTestnet(ctx context.Context) {
	core.SetupLatestBlockHeight(ctx, tonTestnetClient, constant.TON_TESTNET)

	setup.SetupCacheBlockHeight(ctx, constant.TON_TESTNET)

	setup.SetupSweepBlockHeight(ctx, constant.TON_TESTNET)
}

func SweepTonTestnetBlockchainTransaction(ctx context.Context) {
	core.SweepTonBlockchainTransaction(
		ctx,
		tonTestnetClient,
		constant.TON_TESTNET,
		&setup.TonTestnetPublicKey,
		&setup.TonTestnetSweepBlockHeight,
		&setup.TonTestnetCacheBlockHeight,
		constant.TON_TESTNET_SWEEP_BLOCK,
		constant.TON_TESTNET_PENDING_BLOCK,
		constant.TON_TESTNET_PENDING_TRANSACTION)
}

func SweepTonTestnetBlockchainTransactionDetails(ctx context.Context) {
	core.SweepTonBlockchainTransactionDetails(
		ctx,
		tonTestnetClient,
		constant.TON_TESTNET,
		&setup.TonTestnetPublicKey,
		&tonTestnetTxRetryCount,
		constant.TON_TESTNET_PENDING_TRANSACTION)
}

func SweepTonTestnetBlockchainPendingBlock(ctx context.Context) {
	core.SweepTonBlockchainPendingBlock(
		ctx,
		tonTestnetClient,
		constant.TON_TESTNET,
		&setup.TonTestnetPublicKey,
		&tonTestnetBlockRetryCount,
		constant.TON_TESTNET_PENDING_BLOCK,
		constant.TON_TESTNET_PENDING_TRANSACTION)
}
