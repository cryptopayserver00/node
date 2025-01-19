package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	tonSweepCount = make(map[int64]int)

	tonClient NODE_Client.Client
)

func SweepTonBlockchain() {
	initTon()

	go func() {
		for {
			SweepTonBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepTonBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepTonBlockchainPendingBlock()
		}
	}()
}

func initTon() {
	core.SetupTonLatestBlockHeight(tonClient, constant.TON_MAINNET)

	setup.SetupCacheBlockHeight(context.Background(), constant.TON_MAINNET)

	setup.SetupSweepBlockHeight(context.Background(), constant.TON_MAINNET)
}

func SweepTonBlockchainTransaction() {

	core.SweepTonBlockchainTransaction(
		tonClient,
		constant.TON_MAINNET,
		&setup.TonPublicKey,
		&tonSweepCount,
		&setup.TonSweepBlockHeight,
		&setup.TonCacheBlockHeight,
		constant.TON_SWEEP_BLOCK,
		constant.TON_PENDING_BLOCK,
		constant.TON_PENDING_TRANSACTION)
}

func SweepTonBlockchainTransactionDetails() {

	core.SweepTonBlockchainTransactionDetails(
		tonClient,
		constant.TON_MAINNET,
		&setup.TonPublicKey,
		constant.TON_PENDING_TRANSACTION)
}

func SweepTonBlockchainPendingBlock() {
	core.SweepTonBlockchainPendingBlock(
		tonClient,
		constant.TON_MAINNET,
		&setup.TonPublicKey,
		constant.TON_PENDING_BLOCK,
		constant.TON_PENDING_TRANSACTION)
}
