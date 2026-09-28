package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
)

var (
	solBlockRetryCount = make(map[int64]int)
	solTxRetryCount    = make(map[string]int)
)

func SweepSolBlockchain(ctx context.Context) {
	initSol(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepSolBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepSolBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepSolBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initSol(ctx context.Context) {
	core.SetupSolLatestBlockHeight(ctx, constant.SOL_MAINNET)

	setup.SetupCacheBlockHeight(ctx, constant.SOL_MAINNET)

	setup.SetupSweepBlockHeight(ctx, constant.SOL_MAINNET)
}

func SweepSolBlockchainTransaction(ctx context.Context) {
	core.SweepSolBlockchainTransaction(
		ctx,
		constant.SOL_MAINNET,
		&setup.SolPublicKey,
		&setup.SolSweepBlockHeight,
		&setup.SolCacheBlockHeight,
		constant.SOL_SWEEP_BLOCK,
		constant.SOL_PENDING_BLOCK,
		constant.SOL_PENDING_TRANSACTION)
}

func SweepSolBlockchainTransactionDetails(ctx context.Context) {
	core.SweepSolBlockchainTransactionDetails(
		ctx,
		constant.SOL_MAINNET,
		&setup.SolPublicKey,
		&solTxRetryCount,
		constant.SOL_PENDING_TRANSACTION)
}

func SweepSolBlockchainPendingBlock(ctx context.Context) {

	core.SweepSolBlockchainPendingBlock(
		ctx,
		constant.SOL_MAINNET,
		&setup.SolPublicKey,
		&solBlockRetryCount,
		constant.SOL_PENDING_BLOCK,
		constant.SOL_PENDING_TRANSACTION)
}
