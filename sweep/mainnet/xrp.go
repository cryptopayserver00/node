package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	xrpSweepCount = make(map[int64]int)

	xrpClient NODE_Client.Client
)

func SweepXrpBlockchain() {
	initXrp()

	go func() {
		for {
			SweepXrpBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepXrpBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepXrpBlockchainPendingBlock()
		}
	}()
}

func initXrp() {
	core.SetupXrpLatestBlockHeight(xrpClient, constant.XRP_MAINNET)

	setup.SetupCacheBlockHeight(context.Background(), constant.XRP_MAINNET)

	setup.SetupSweepBlockHeight(context.Background(), constant.XRP_MAINNET)
}

func SweepXrpBlockchainTransaction() {

	core.SweepXrpBlockchainTransaction(
		xrpClient,
		constant.XRP_MAINNET,
		&setup.XrpPublicKey,
		&xrpSweepCount,
		&setup.XrpSweepBlockHeight,
		&setup.XrpCacheBlockHeight,
		constant.XRP_SWEEP_BLOCK,
		constant.XRP_PENDING_BLOCK,
		constant.XRP_PENDING_TRANSACTION)
}

func SweepXrpBlockchainTransactionDetails() {

	core.SweepXrpBlockchainTransactionDetails(
		xrpClient,
		constant.XRP_MAINNET,
		&setup.XrpPublicKey,
		constant.XRP_PENDING_TRANSACTION)
}

func SweepXrpBlockchainPendingBlock() {
	core.SweepXrpBlockchainPendingBlock(
		xrpClient,
		constant.XRP_MAINNET,
		&setup.XrpPublicKey,
		constant.XRP_PENDING_BLOCK,
		constant.XRP_PENDING_TRANSACTION)
}
