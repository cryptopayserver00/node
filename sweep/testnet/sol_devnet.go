package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	solDevnetSweepCount = make(map[int64]int)

	solDevnetClient NODE_Client.Client
)

func SweepSolDevnetBlockchain() {
	initSolDevnet()

	go func() {
		for {
			SweepSolDevnetBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepSolDevnetBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepSolDevnetBlockchainPendingBlock()
		}
	}()
}

func initSolDevnet() {
	core.SetupLatestBlockHeight(solDevnetClient, constant.SOL_DEVNET)

	setup.SetupCacheBlockHeight(context.Background(), constant.SOL_DEVNET)

	setup.SetupSweepBlockHeight(context.Background(), constant.SOL_DEVNET)
}

func SweepSolDevnetBlockchainTransaction() {

	core.SweepSolBlockchainTransaction(
		solDevnetClient,
		constant.SOL_DEVNET,
		&setup.SolDevnetPublicKey,
		&solDevnetSweepCount,
		&setup.SolDevnetSweepBlockHeight,
		&setup.SolDevnetCacheBlockHeight,
		constant.SOL_DEVNET_SWEEP_BLOCK,
		constant.SOL_DEVNET_PENDING_BLOCK,
		constant.SOL_DEVNET_PENDING_TRANSACTION)
}

func SweepSolDevnetBlockchainTransactionDetails() {

	core.SweepSolBlockchainTransactionDetails(
		solDevnetClient,
		constant.SOL_DEVNET,
		&setup.SolDevnetPublicKey,
		constant.SOL_DEVNET_PENDING_TRANSACTION)
}

func SweepSolDevnetBlockchainPendingBlock() {
	core.SweepSolBlockchainPendingBlock(
		solDevnetClient,
		constant.SOL_DEVNET,
		&setup.SolDevnetPublicKey,
		constant.SOL_DEVNET_PENDING_BLOCK,
		constant.SOL_DEVNET_PENDING_TRANSACTION)
}
