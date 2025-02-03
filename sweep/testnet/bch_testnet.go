package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	bchTestnetSweepCount = make(map[int64]int)

	bchTestnetClient NODE_Client.Client
)

func SweepBchTestnetBlockchain() {
	initBchTestnet()

	go func() {
		for {
			SweepBchTestnetBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepBchTestnetBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepBchTestnetBlockchainPendingBlock()
		}
	}()
}

func initBchTestnet() {
	core.SetupLatestBlockHeight(bchTestnetClient, constant.BCH_TESTNET)

	setup.SetupCacheBlockHeight(context.Background(), constant.BCH_TESTNET)

	setup.SetupSweepBlockHeight(context.Background(), constant.BCH_TESTNET)
}

func SweepBchTestnetBlockchainTransaction() {

	core.SweepBchBlockchainTransaction(
		bchTestnetClient,
		constant.BCH_TESTNET,
		&setup.BchTestnetPublicKey,
		&bchTestnetSweepCount,
		&setup.BchTestnetSweepBlockHeight,
		&setup.BchTestnetCacheBlockHeight,
		constant.BCH_TESTNET_SWEEP_BLOCK,
		constant.BCH_TESTNET_PENDING_BLOCK,
		constant.BCH_TESTNET_PENDING_TRANSACTION)
}

func SweepBchTestnetBlockchainTransactionDetails() {

	core.SweepBchBlockchainTransactionDetails(
		bchTestnetClient,
		constant.BCH_TESTNET,
		&setup.BchTestnetPublicKey,
		constant.BCH_TESTNET_PENDING_TRANSACTION)
}

func SweepBchTestnetBlockchainPendingBlock() {
	core.SweepBchBlockchainPendingBlock(
		bchTestnetClient,
		constant.BCH_TESTNET,
		&setup.BchTestnetPublicKey,
		constant.BCH_TESTNET_PENDING_BLOCK,
		constant.BCH_TESTNET_PENDING_TRANSACTION)
}
