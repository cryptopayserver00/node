package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	ethSepoliaBlockRetryCount = make(map[int64]int)
	ethSepoliaTxRetryCount    = make(map[string]int)

	ethSepoliaClient NODE_Client.Client
)

func SweepEthSepoliaBlockchain(ctx context.Context) {
	initEthSepolia(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepEthSepoliaBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepEthSepoliaBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepEthSepoliaBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initEthSepolia(ctx context.Context) {
	core.SetupLatestBlockHeight(ctx, ethSepoliaClient, constant.ETH_SEPOLIA)

	setup.SetupCacheBlockHeight(ctx, constant.ETH_SEPOLIA)

	setup.SetupSweepBlockHeight(ctx, constant.ETH_SEPOLIA)
}

func SweepEthSepoliaBlockchainTransaction(ctx context.Context) {
	core.SweepBlockchainTransaction(
		ctx,
		ethSepoliaClient,
		constant.ETH_SEPOLIA,
		&setup.EthSepoliaPublicKey,
		&setup.EthSepoliaSweepBlockHeight,
		&setup.EthSepoliaCacheBlockHeight,
		constant.ETH_SEPOLIA_SWEEP_BLOCK,
		constant.ETH_SEPOLIA_PENDING_BLOCK,
		constant.ETH_SEPOLIA_PENDING_TRANSACTION)
}

func SweepEthSepoliaBlockchainTransactionDetails(ctx context.Context) {

	core.SweepBlockchainTransactionDetails(
		ctx,
		ethSepoliaClient,
		constant.ETH_SEPOLIA,
		&setup.EthSepoliaPublicKey,
		&ethSepoliaTxRetryCount,
		constant.ETH_SEPOLIA_PENDING_TRANSACTION)
}

func SweepEthSepoliaBlockchainPendingBlock(ctx context.Context) {

	core.SweepBlockchainPendingBlock(
		ctx,
		ethSepoliaClient,
		constant.ETH_SEPOLIA,
		&setup.EthSepoliaPublicKey,
		&ethSepoliaBlockRetryCount,
		constant.ETH_SEPOLIA_PENDING_BLOCK,
		constant.ETH_SEPOLIA_PENDING_TRANSACTION)
}
