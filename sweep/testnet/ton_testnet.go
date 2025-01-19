package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	tonTestnetSweepCount = make(map[int64]int)

	tonTestnetClient NODE_Client.Client
)

func SweepTonTestnetBlockchain() {
	initTonTestnet()

	go func() {
		for {
			SweepTonTestnetBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepTonTestnetBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepTonTestnetBlockchainPendingBlock()
		}
	}()
}

func initTonTestnet() {
	core.SetupLatestBlockHeight(tonTestnetClient, constant.TON_TESTNET)

	setup.SetupCacheBlockHeight(context.Background(), constant.TON_TESTNET)

	setup.SetupSweepBlockHeight(context.Background(), constant.TON_TESTNET)
}

func SweepTonTestnetBlockchainTransaction() {

	core.SweepTonBlockchainTransaction(
		tonTestnetClient,
		constant.TON_TESTNET,
		&setup.TonTestnetPublicKey,
		&tonTestnetSweepCount,
		&setup.TonTestnetSweepBlockHeight,
		&setup.TonTestnetCacheBlockHeight,
		constant.TON_TESTNET_SWEEP_BLOCK,
		constant.TON_TESTNET_PENDING_BLOCK,
		constant.TON_TESTNET_PENDING_TRANSACTION)
}

func SweepTonTestnetBlockchainTransactionDetails() {

	core.SweepTonBlockchainTransactionDetails(
		tonTestnetClient,
		constant.TON_TESTNET,
		&setup.TonTestnetPublicKey,
		constant.TON_TESTNET_PENDING_TRANSACTION)
}

func SweepTonTestnetBlockchainPendingBlock() {
	core.SweepTonBlockchainPendingBlock(
		tonTestnetClient,
		constant.TON_TESTNET,
		&setup.TonTestnetPublicKey,
		constant.TON_TESTNET_PENDING_BLOCK,
		constant.TON_TESTNET_PENDING_TRANSACTION)
}
