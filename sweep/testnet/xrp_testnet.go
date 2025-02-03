package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	xrpTestnetSweepCount = make(map[int64]int)

	xrpTestnetClient NODE_Client.Client
)

func SweepXrpTestnetBlockchain() {
	initXrpTestnet()

	go func() {
		for {
			SweepXrpTestnetBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepXrpTestnetBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepXrpTestnetBlockchainPendingBlock()
		}
	}()
}

func initXrpTestnet() {
	core.SetupLatestBlockHeight(xrpTestnetClient, constant.XRP_TESTNET)

	setup.SetupCacheBlockHeight(context.Background(), constant.XRP_TESTNET)

	setup.SetupSweepBlockHeight(context.Background(), constant.XRP_TESTNET)
}

func SweepXrpTestnetBlockchainTransaction() {

	core.SweepXrpBlockchainTransaction(
		xrpTestnetClient,
		constant.XRP_TESTNET,
		&setup.XrpTestnetPublicKey,
		&xrpTestnetSweepCount,
		&setup.XrpTestnetSweepBlockHeight,
		&setup.XrpTestnetCacheBlockHeight,
		constant.XRP_TESTNET_SWEEP_BLOCK,
		constant.XRP_TESTNET_PENDING_BLOCK,
		constant.XRP_TESTNET_PENDING_TRANSACTION)
}

func SweepXrpTestnetBlockchainTransactionDetails() {

	core.SweepXrpBlockchainTransactionDetails(
		xrpTestnetClient,
		constant.XRP_TESTNET,
		&setup.XrpTestnetPublicKey,
		constant.XRP_TESTNET_PENDING_TRANSACTION)
}

func SweepXrpTestnetBlockchainPendingBlock() {
	core.SweepXrpBlockchainPendingBlock(
		xrpTestnetClient,
		constant.XRP_TESTNET,
		&setup.XrpTestnetPublicKey,
		constant.XRP_TESTNET_PENDING_BLOCK,
		constant.XRP_TESTNET_PENDING_TRANSACTION)
}
