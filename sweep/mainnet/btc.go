package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	btcSweepCount = make(map[int64]int)

	btcClient NODE_Client.Client
)

func SweepBtcBlockchain() {
	initBtc()

	go func() {
		for {
			SweepBtcBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepBtcBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepBtcBlockchainPendingBlock()
		}
	}()
}

func initBtc() {
	core.SetupBtcLatestBlockHeight(btcClient, constant.BTC_MAINNET)

	setup.SetupCacheBlockHeight(context.Background(), constant.BTC_MAINNET)

	setup.SetupSweepBlockHeight(context.Background(), constant.BTC_MAINNET)
}

func SweepBtcBlockchainTransaction() {

	core.SweepBtcBlockchainTransaction(
		btcClient,
		constant.BTC_MAINNET,
		&setup.BtcPublicKey,
		&btcSweepCount,
		&setup.BtcSweepBlockHeight,
		&setup.BtcCacheBlockHeight,
		constant.BTC_SWEEP_BLOCK,
		constant.BTC_PENDING_BLOCK,
		constant.BTC_PENDING_TRANSACTION)
}

func SweepBtcBlockchainTransactionDetails() {

	core.SweepBtcBlockchainTransactionDetails(
		btcClient,
		constant.BTC_MAINNET,
		&setup.BtcPublicKey,
		constant.BTC_PENDING_TRANSACTION)
}

func SweepBtcBlockchainPendingBlock() {
	core.SweepBtcBlockchainPendingBlock(
		btcClient,
		constant.BTC_MAINNET,
		&setup.BtcPublicKey,
		constant.BTC_PENDING_BLOCK,
		constant.BTC_PENDING_TRANSACTION)
}
