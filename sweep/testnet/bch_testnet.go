package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	bchTestnetBlockRetryCount = make(map[int64]int)
	bchTestnetTxRetryCount    = make(map[string]int)

	bchTestnetClient NODE_Client.Client
)

func SweepBchTestnetBlockchain(ctx context.Context) {
	initBchTestnet(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBchTestnetBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBchTestnetBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBchTestnetBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initBchTestnet(ctx context.Context) {
	core.SetupLatestBlockHeight(ctx, bchTestnetClient, constant.BCH_TESTNET)

	setup.SetupCacheBlockHeight(ctx, constant.BCH_TESTNET)

	setup.SetupSweepBlockHeight(ctx, constant.BCH_TESTNET)
}

func SweepBchTestnetBlockchainTransaction(ctx context.Context) {
	core.SweepBchBlockchainTransaction(
		ctx,
		bchTestnetClient,
		constant.BCH_TESTNET,
		&setup.BchTestnetPublicKey,
		&setup.BchTestnetSweepBlockHeight,
		&setup.BchTestnetCacheBlockHeight,
		constant.BCH_TESTNET_SWEEP_BLOCK,
		constant.BCH_TESTNET_PENDING_BLOCK,
		constant.BCH_TESTNET_PENDING_TRANSACTION)
}

func SweepBchTestnetBlockchainTransactionDetails(ctx context.Context) {

	core.SweepBchBlockchainTransactionDetails(
		ctx,
		bchTestnetClient,
		constant.BCH_TESTNET,
		&setup.BchTestnetPublicKey,
		&bchTestnetTxRetryCount,
		constant.BCH_TESTNET_PENDING_TRANSACTION)
}

func SweepBchTestnetBlockchainPendingBlock(ctx context.Context) {
	core.SweepBchBlockchainPendingBlock(
		ctx,
		bchTestnetClient,
		constant.BCH_TESTNET,
		&setup.BchTestnetPublicKey,
		&bchTestnetBlockRetryCount,
		constant.BCH_TESTNET_PENDING_BLOCK,
		constant.BCH_TESTNET_PENDING_TRANSACTION)
}
