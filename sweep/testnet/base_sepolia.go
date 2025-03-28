package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	baseSepoliaSweepCount = make(map[int64]int)

	baseSepoliaClient NODE_Client.Client
)

func SweepBaseSepoliaBlockchain() {
	initBaseSepolia()

	go func() {
		for {
			SweepBaseSepoliaBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepBaseSepoliaBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepBaseSepoliaBlockchainPendingBlock()
		}
	}()
}

func initBaseSepolia() {
	core.SetupLatestBlockHeight(baseSepoliaClient, constant.BASE_SEPOLIA)

	setup.SetupCacheBlockHeight(context.Background(), constant.BASE_SEPOLIA)

	setup.SetupSweepBlockHeight(context.Background(), constant.BASE_SEPOLIA)
}

func SweepBaseSepoliaBlockchainTransaction() {

	core.SweepBlockchainTransaction(
		baseSepoliaClient,
		constant.BASE_SEPOLIA,
		&setup.BaseSepoliaPublicKey,
		&baseSepoliaSweepCount,
		&setup.BaseSepoliaSweepBlockHeight,
		&setup.BaseSepoliaCacheBlockHeight,
		constant.BASE_SEPOLIA_SWEEP_BLOCK,
		constant.BASE_SEPOLIA_PENDING_BLOCK,
		constant.BASE_SEPOLIA_PENDING_TRANSACTION)
}

func SweepBaseSepoliaBlockchainTransactionDetails() {

	core.SweepBlockchainTransactionDetails(
		baseSepoliaClient,
		constant.BASE_SEPOLIA,
		&setup.BaseSepoliaPublicKey,
		constant.BASE_SEPOLIA_PENDING_TRANSACTION)
}

func SweepBaseSepoliaBlockchainPendingBlock() {
	core.SweepBlockchainPendingBlock(
		baseSepoliaClient,
		constant.BASE_SEPOLIA,
		&setup.BaseSepoliaPublicKey,
		&baseSepoliaSweepCount,
		constant.BASE_SEPOLIA_PENDING_BLOCK,
		constant.BASE_SEPOLIA_PENDING_TRANSACTION)
}
