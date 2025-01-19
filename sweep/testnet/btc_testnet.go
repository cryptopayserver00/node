package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	btcTestnetSweepCount = make(map[int64]int)

	btcTestnetClient NODE_Client.Client
)

func SweepBtcTestnetBlockchain() {
	initBtcTestnet()

	go func() {
		for {
			SweepBtcTestnetBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepBtcTestnetBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepBtcTestnetBlockchainPendingBlock()
		}
	}()
}

func initBtcTestnet() {
	core.SetupBtcLatestBlockHeight(btcTestnetClient, constant.BTC_TESTNET)

	setup.SetupCacheBlockHeight(context.Background(), constant.BTC_TESTNET)

	setup.SetupSweepBlockHeight(context.Background(), constant.BTC_TESTNET)
}

func SweepBtcTestnetBlockchainTransaction() {
	core.SweepBtcBlockchainTransaction(
		btcTestnetClient,
		constant.BTC_TESTNET,
		&setup.BtcTestnetPublicKey,
		&btcTestnetSweepCount,
		&setup.BtcTestnetSweepBlockHeight,
		&setup.BtcTestnetCacheBlockHeight,
		constant.BTC_TESTNET_SWEEP_BLOCK,
		constant.BTC_TESTNET_PENDING_BLOCK,
		constant.BTC_TESTNET_PENDING_TRANSACTION)
}

func SweepBtcTestnetBlockchainTransactionDetails() {
	core.SweepBtcBlockchainTransactionDetails(
		btcTestnetClient,
		constant.BTC_TESTNET,
		&setup.BtcTestnetPublicKey,
		constant.BTC_TESTNET_PENDING_TRANSACTION)
}

func SweepBtcTestnetBlockchainPendingBlock() {
	core.SweepBtcBlockchainPendingBlock(
		btcTestnetClient,
		constant.BTC_TESTNET,
		&setup.BtcTestnetPublicKey,
		constant.BTC_TESTNET_PENDING_BLOCK,
		constant.BTC_TESTNET_PENDING_TRANSACTION)
}
