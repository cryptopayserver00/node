package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	baseSepoliaBlockRetryCount = make(map[int64]int)
	baseSepoliaTxRetryCount    = make(map[string]int)

	baseSepoliaClient NODE_Client.Client
)

func SweepBaseSepoliaBlockchain(ctx context.Context) {
	initBaseSepolia(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBaseSepoliaBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBaseSepoliaBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBaseSepoliaBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initBaseSepolia(ctx context.Context) {
	core.SetupLatestBlockHeight(ctx, baseSepoliaClient, constant.BASE_SEPOLIA)

	setup.SetupCacheBlockHeight(ctx, constant.BASE_SEPOLIA)

	setup.SetupSweepBlockHeight(ctx, constant.BASE_SEPOLIA)
}

func SweepBaseSepoliaBlockchainTransaction(ctx context.Context) {

	core.SweepBlockchainTransaction(
		ctx,
		baseSepoliaClient,
		constant.BASE_SEPOLIA,
		&setup.BaseSepoliaPublicKey,
		&setup.BaseSepoliaSweepBlockHeight,
		&setup.BaseSepoliaCacheBlockHeight,
		constant.BASE_SEPOLIA_SWEEP_BLOCK,
		constant.BASE_SEPOLIA_PENDING_BLOCK,
		constant.BASE_SEPOLIA_PENDING_TRANSACTION)
}

func SweepBaseSepoliaBlockchainTransactionDetails(ctx context.Context) {

	core.SweepBlockchainTransactionDetails(
		ctx,
		baseSepoliaClient,
		constant.BASE_SEPOLIA,
		&setup.BaseSepoliaPublicKey,
		&baseSepoliaTxRetryCount,
		constant.BASE_SEPOLIA_PENDING_TRANSACTION)
}

func SweepBaseSepoliaBlockchainPendingBlock(ctx context.Context) {
	core.SweepBlockchainPendingBlock(
		ctx,
		baseSepoliaClient,
		constant.BASE_SEPOLIA,
		&setup.BaseSepoliaPublicKey,
		&baseSepoliaBlockRetryCount,
		constant.BASE_SEPOLIA_PENDING_BLOCK,
		constant.BASE_SEPOLIA_PENDING_TRANSACTION)
}
