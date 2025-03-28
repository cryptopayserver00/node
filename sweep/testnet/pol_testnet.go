package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	polTestnetSweepCount = make(map[int64]int)

	polTestnetClient NODE_Client.Client
)

func SweepPolTestnetBlockchain() {
	initPolTestnet()

	go func() {
		for {
			SweepPolTestnetBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepPolTestnetBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepPolTestnetBlockchainPendingBlock()
		}
	}()
}

func initPolTestnet() {
	core.SetupLatestBlockHeight(polTestnetClient, constant.POL_TESTNET)

	setup.SetupCacheBlockHeight(context.Background(), constant.POL_TESTNET)

	setup.SetupSweepBlockHeight(context.Background(), constant.POL_TESTNET)
}

func SweepPolTestnetBlockchainTransaction() {

	core.SweepBlockchainTransaction(
		polTestnetClient,
		constant.POL_TESTNET,
		&setup.PolTestnetPublicKey,
		&polTestnetSweepCount,
		&setup.PolTestnetSweepBlockHeight,
		&setup.PolTestnetCacheBlockHeight,
		constant.POL_TESTNET_SWEEP_BLOCK,
		constant.POL_TESTNET_PENDING_BLOCK,
		constant.POL_TESTNET_PENDING_TRANSACTION)
}

func SweepPolTestnetBlockchainTransactionDetails() {

	core.SweepBlockchainTransactionDetails(
		polTestnetClient,
		constant.POL_TESTNET,
		&setup.PolTestnetPublicKey,
		constant.POL_TESTNET_PENDING_TRANSACTION)
}

func SweepPolTestnetBlockchainPendingBlock() {
	core.SweepBlockchainPendingBlock(
		polTestnetClient,
		constant.POL_TESTNET,
		&setup.PolTestnetPublicKey,
		&polTestnetSweepCount,
		constant.POL_TESTNET_PENDING_BLOCK,
		constant.POL_TESTNET_PENDING_TRANSACTION)
}
