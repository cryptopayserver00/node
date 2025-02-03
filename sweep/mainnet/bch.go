package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	bchSweepCount = make(map[int64]int)

	bchClient NODE_Client.Client
)

func SweepBchBlockchain() {
	initBch()

	go func() {
		for {
			SweepBchBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepBchBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepBchBlockchainPendingBlock()
		}
	}()
}

func initBch() {
	core.SetupBchLatestBlockHeight(bchClient, constant.BCH_MAINNET)

	setup.SetupCacheBlockHeight(context.Background(), constant.BCH_MAINNET)

	setup.SetupSweepBlockHeight(context.Background(), constant.BCH_MAINNET)
}

func SweepBchBlockchainTransaction() {

	core.SweepBchBlockchainTransaction(
		bchClient,
		constant.BCH_MAINNET,
		&setup.BchPublicKey,
		&bchSweepCount,
		&setup.BchSweepBlockHeight,
		&setup.BchCacheBlockHeight,
		constant.BCH_SWEEP_BLOCK,
		constant.BCH_PENDING_BLOCK,
		constant.BCH_PENDING_TRANSACTION)
}

func SweepBchBlockchainTransactionDetails() {

	core.SweepBchBlockchainTransactionDetails(
		bchClient,
		constant.BCH_MAINNET,
		&setup.BchPublicKey,
		constant.BCH_PENDING_TRANSACTION)
}

func SweepBchBlockchainPendingBlock() {
	core.SweepBchBlockchainPendingBlock(
		bchClient,
		constant.BCH_MAINNET,
		&setup.BchPublicKey,
		constant.BCH_PENDING_BLOCK,
		constant.BCH_PENDING_TRANSACTION)
}
