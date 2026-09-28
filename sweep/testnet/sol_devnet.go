package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
)

var (
	solDevnetBlockRetryCount = make(map[int64]int)
	solDevnetTxRetryCount    = make(map[string]int)
)

func SweepSolDevnetBlockchain(ctx context.Context) {
	initSolDevnet(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepSolDevnetBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepSolDevnetBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepSolDevnetBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initSolDevnet(ctx context.Context) {
	core.SetupSolLatestBlockHeight(ctx, constant.SOL_DEVNET)

	setup.SetupCacheBlockHeight(ctx, constant.SOL_DEVNET)

	setup.SetupSweepBlockHeight(ctx, constant.SOL_DEVNET)
}

func SweepSolDevnetBlockchainTransaction(ctx context.Context) {
	core.SweepSolBlockchainTransaction(
		ctx,
		constant.SOL_DEVNET,
		&setup.SolDevnetPublicKey,
		&setup.SolDevnetSweepBlockHeight,
		&setup.SolDevnetCacheBlockHeight,
		constant.SOL_DEVNET_SWEEP_BLOCK,
		constant.SOL_DEVNET_PENDING_BLOCK,
		constant.SOL_DEVNET_PENDING_TRANSACTION)
}

func SweepSolDevnetBlockchainTransactionDetails(ctx context.Context) {
	core.SweepSolBlockchainTransactionDetails(
		ctx,
		constant.SOL_DEVNET,
		&setup.SolDevnetPublicKey,
		&solDevnetTxRetryCount,
		constant.SOL_DEVNET_PENDING_TRANSACTION)
}

func SweepSolDevnetBlockchainPendingBlock(ctx context.Context) {
	core.SweepSolBlockchainPendingBlock(
		ctx,
		constant.SOL_DEVNET,
		&setup.SolDevnetPublicKey,
		&solDevnetBlockRetryCount,
		constant.SOL_DEVNET_PENDING_BLOCK,
		constant.SOL_DEVNET_PENDING_TRANSACTION)
}
