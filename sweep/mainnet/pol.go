package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	polSweepCount = make(map[int64]int)

	polClient NODE_Client.Client
)

func SweepPolBlockchain() {
	initPol()

	go func() {
		for {
			SweepPolBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepPolBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepPolBlockchainPendingBlock()
		}
	}()
}

func initPol() {
	core.SetupLatestBlockHeight(polClient, constant.POL_MAINNET)

	setup.SetupCacheBlockHeight(context.Background(), constant.POL_MAINNET)

	setup.SetupSweepBlockHeight(context.Background(), constant.POL_MAINNET)
}

func SweepPolBlockchainTransaction() {

	core.SweepBlockchainTransaction(
		polClient,
		constant.POL_MAINNET,
		&setup.PolPublicKey,
		&polSweepCount,
		&setup.PolSweepBlockHeight,
		&setup.PolCacheBlockHeight,
		constant.POL_SWEEP_BLOCK,
		constant.POL_PENDING_BLOCK,
		constant.POL_PENDING_TRANSACTION)
}

func SweepPolBlockchainTransactionDetails() {

	core.SweepBlockchainTransactionDetails(
		polClient,
		constant.POL_MAINNET,
		&setup.PolPublicKey,
		constant.POL_PENDING_TRANSACTION)
}

func SweepPolBlockchainPendingBlock() {
	core.SweepBlockchainPendingBlock(
		polClient,
		constant.POL_MAINNET,
		&setup.PolPublicKey,
		&polSweepCount,
		constant.POL_PENDING_BLOCK,
		constant.POL_PENDING_TRANSACTION)
}
