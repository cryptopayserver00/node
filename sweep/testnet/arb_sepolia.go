package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	arbitrumSepoliaSweepCount = make(map[int64]int)

	arbitrumSepoliaClient NODE_Client.Client
)

func SweepArbitrumSepoliaBlockchain() {

	initArbitrumSepolia()

	go func() {
		for {
			SweepArbitrumSepoliaBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepArbitrumSepoliaBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepArbitrumSepoliaBlockchainPendingBlock()
		}
	}()
}

func initArbitrumSepolia() {
	core.SetupLatestBlockHeight(arbitrumSepoliaClient, constant.ARBITRUM_SEPOLIA)

	setup.SetupCacheBlockHeight(context.Background(), constant.ARBITRUM_SEPOLIA)

	setup.SetupSweepBlockHeight(context.Background(), constant.ARBITRUM_SEPOLIA)
}

func SweepArbitrumSepoliaBlockchainTransaction() {
	core.SweepBlockchainTransaction(
		arbitrumSepoliaClient,
		constant.ARBITRUM_SEPOLIA,
		&setup.ArbitrumSepoliaPublicKey,
		&arbitrumSepoliaSweepCount,
		&setup.ArbitrumSepoliaSweepBlockHeight,
		&setup.ArbitrumSepoliaCacheBlockHeight,
		constant.ARBITRUM_SEPOLIA_SWEEP_BLOCK,
		constant.ARBITRUM_SEPOLIA_PENDING_BLOCK,
		constant.ARBITRUM_SEPOLIA_PENDING_TRANSACTION)
}

func SweepArbitrumSepoliaBlockchainTransactionDetails() {
	core.SweepBlockchainTransactionDetails(
		arbitrumSepoliaClient,
		constant.ARBITRUM_SEPOLIA,
		&setup.ArbitrumSepoliaPublicKey,
		constant.ARBITRUM_SEPOLIA_PENDING_TRANSACTION)
}

func SweepArbitrumSepoliaBlockchainPendingBlock() {
	core.SweepBlockchainPendingBlock(
		arbitrumSepoliaClient,
		constant.ARBITRUM_SEPOLIA,
		&setup.ArbitrumSepoliaPublicKey,
		&arbitrumSepoliaSweepCount,
		constant.ARBITRUM_SEPOLIA_PENDING_BLOCK,
		constant.ARBITRUM_SEPOLIA_PENDING_TRANSACTION)
}
