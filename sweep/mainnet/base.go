package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	baseSweepCount = make(map[int64]int)

	baseClient NODE_Client.Client
)

func SweepBaseBlockchain() {
	initBase()

	go func() {
		for {
			SweepBaseBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepBaseBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepBaseBlockchainPendingBlock()
		}
	}()
}

func initBase() {
	core.SetupLatestBlockHeight(baseClient, constant.BASE_MAINNET)

	setup.SetupCacheBlockHeight(context.Background(), constant.BASE_MAINNET)

	setup.SetupSweepBlockHeight(context.Background(), constant.BASE_MAINNET)
}

func SweepBaseBlockchainTransaction() {

	core.SweepBlockchainTransaction(
		baseClient,
		constant.BASE_MAINNET,
		&setup.BasePublicKey,
		&baseSweepCount,
		&setup.BaseSweepBlockHeight,
		&setup.BaseCacheBlockHeight,
		constant.BASE_SWEEP_BLOCK,
		constant.BASE_PENDING_BLOCK,
		constant.BASE_PENDING_TRANSACTION)
}

func SweepBaseBlockchainTransactionDetails() {

	core.SweepBlockchainTransactionDetails(
		baseClient,
		constant.BASE_MAINNET,
		&setup.BasePublicKey,
		constant.BASE_PENDING_TRANSACTION)
}

func SweepBaseBlockchainPendingBlock() {
	core.SweepBlockchainPendingBlock(
		baseClient,
		constant.BASE_MAINNET,
		&setup.BasePublicKey,
		&baseSweepCount,
		constant.BASE_PENDING_BLOCK,
		constant.BASE_PENDING_TRANSACTION)
}
