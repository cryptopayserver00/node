package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	tronNileSweepCount = make(map[int64]int)

	tronNileClient NODE_Client.Client
)

func SweepTronNileBlockchain() {
	initTronNile()

	go func() {
		for {
			SweepTronNileBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepTronNileBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepTronNileBlockchainPendingBlock()
		}
	}()
}

func initTronNile() {
	core.SetupTronLatestBlockHeight(tronNileClient, constant.TRON_NILE)

	setup.SetupCacheBlockHeight(context.Background(), constant.TRON_NILE)

	setup.SetupSweepBlockHeight(context.Background(), constant.TRON_NILE)
}

func SweepTronNileBlockchainTransaction() {
	core.SweepTronBlockchainTransaction(
		tronNileClient,
		constant.TRON_NILE,
		&setup.TronNilePublicKey,
		&tronNileSweepCount,
		&setup.TronNileSweepBlockHeight,
		&setup.TronNileCacheBlockHeight,
		constant.TRON_NILE_SWEEP_BLOCK,
		constant.TRON_NILE_PENDING_BLOCK,
		constant.TRON_NILE_PENDING_TRANSACTION)
}

func SweepTronNileBlockchainTransactionDetails() {
	core.SweepTronBlockchainTransactionDetails(
		tronNileClient,
		constant.TRON_NILE,
		&setup.TronNilePublicKey,
		constant.TRON_NILE_PENDING_TRANSACTION)
}

func SweepTronNileBlockchainPendingBlock() {
	core.SweepTronBlockchainPendingBlock(
		tronNileClient,
		constant.TRON_NILE,
		&setup.TronNilePublicKey,
		constant.TRON_NILE_PENDING_BLOCK,
		constant.TRON_NILE_PENDING_TRANSACTION)
}
