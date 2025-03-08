package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
)

var (
	solSweepCount = make(map[int64]int)
)

func SweepSolBlockchain() {
	initSol()

	go func() {
		for {
			SweepSolBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepSolBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepSolBlockchainPendingBlock()
		}
	}()
}

func initSol() {
	core.SetupSolLatestBlockHeight(constant.SOL_MAINNET)

	setup.SetupCacheBlockHeight(context.Background(), constant.SOL_MAINNET)

	setup.SetupSweepBlockHeight(context.Background(), constant.SOL_MAINNET)
}

func SweepSolBlockchainTransaction() {

	core.SweepSolBlockchainTransaction(
		constant.SOL_MAINNET,
		&setup.SolPublicKey,
		&solSweepCount,
		&setup.SolSweepBlockHeight,
		&setup.SolCacheBlockHeight,
		constant.SOL_SWEEP_BLOCK,
		constant.SOL_PENDING_BLOCK,
		constant.SOL_PENDING_TRANSACTION)
}

func SweepSolBlockchainTransactionDetails() {

	core.SweepSolBlockchainTransactionDetails(
		constant.SOL_MAINNET,
		&setup.SolPublicKey,
		constant.SOL_PENDING_TRANSACTION)
}

func SweepSolBlockchainPendingBlock() {

	core.SweepSolBlockchainPendingBlock(
		constant.SOL_MAINNET,
		&setup.SolPublicKey,
		constant.SOL_PENDING_BLOCK,
		constant.SOL_PENDING_TRANSACTION)
}
