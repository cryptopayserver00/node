package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	arbitrumSepoliaBlockRetryCount = make(map[int64]int)
	arbitrumSepoliaTxRetryCount    = make(map[string]int)

	arbitrumSepoliaClient NODE_Client.Client
)

func SweepArbitrumSepoliaBlockchain(ctx context.Context) {

	initArbitrumSepolia(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepArbitrumSepoliaBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepArbitrumSepoliaBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepArbitrumSepoliaBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initArbitrumSepolia(ctx context.Context) {
	core.SetupLatestBlockHeight(ctx, arbitrumSepoliaClient, constant.ARBITRUM_SEPOLIA)

	setup.SetupCacheBlockHeight(ctx, constant.ARBITRUM_SEPOLIA)

	setup.SetupSweepBlockHeight(ctx, constant.ARBITRUM_SEPOLIA)
}

func SweepArbitrumSepoliaBlockchainTransaction(ctx context.Context) {
	core.SweepBlockchainTransaction(
		ctx,
		arbitrumSepoliaClient,
		constant.ARBITRUM_SEPOLIA,
		&setup.ArbitrumSepoliaPublicKey,
		&setup.ArbitrumSepoliaSweepBlockHeight,
		&setup.ArbitrumSepoliaCacheBlockHeight,
		constant.ARBITRUM_SEPOLIA_SWEEP_BLOCK,
		constant.ARBITRUM_SEPOLIA_PENDING_BLOCK,
		constant.ARBITRUM_SEPOLIA_PENDING_TRANSACTION)
}

func SweepArbitrumSepoliaBlockchainTransactionDetails(ctx context.Context) {
	core.SweepBlockchainTransactionDetails(
		ctx,
		arbitrumSepoliaClient,
		constant.ARBITRUM_SEPOLIA,
		&setup.ArbitrumSepoliaPublicKey,
		&arbitrumSepoliaTxRetryCount,
		constant.ARBITRUM_SEPOLIA_PENDING_TRANSACTION)
}

func SweepArbitrumSepoliaBlockchainPendingBlock(ctx context.Context) {
	core.SweepBlockchainPendingBlock(
		ctx,
		arbitrumSepoliaClient,
		constant.ARBITRUM_SEPOLIA,
		&setup.ArbitrumSepoliaPublicKey,
		&arbitrumSepoliaBlockRetryCount,
		constant.ARBITRUM_SEPOLIA_PENDING_BLOCK,
		constant.ARBITRUM_SEPOLIA_PENDING_TRANSACTION)
}
