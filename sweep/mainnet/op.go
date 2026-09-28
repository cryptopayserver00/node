package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	opBlockRetryCount = make(map[int64]int)
	opTxRetryCount    = make(map[string]int)

	opClient NODE_Client.Client
)

func SweepOpBlockchain(ctx context.Context) {

	initOp(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepOpBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepOpBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepOpBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initOp(ctx context.Context) {
	core.SetupLatestBlockHeight(ctx, opClient, constant.OP_MAINNET)

	setup.SetupCacheBlockHeight(ctx, constant.OP_MAINNET)

	setup.SetupSweepBlockHeight(ctx, constant.OP_MAINNET)
}

func SweepOpBlockchainTransaction(ctx context.Context) {
	core.SweepBlockchainTransaction(
		ctx,
		opClient,
		constant.OP_MAINNET,
		&setup.OpPublicKey,
		&setup.OpSweepBlockHeight,
		&setup.OpCacheBlockHeight,
		constant.OP_SWEEP_BLOCK,
		constant.OP_PENDING_BLOCK,
		constant.OP_PENDING_TRANSACTION)
}

func SweepOpBlockchainTransactionDetails(ctx context.Context) {
	core.SweepBlockchainTransactionDetails(
		ctx,
		opClient,
		constant.OP_MAINNET,
		&setup.OpPublicKey,
		&opTxRetryCount,
		constant.OP_PENDING_TRANSACTION)
}

func SweepOpBlockchainPendingBlock(ctx context.Context) {
	core.SweepBlockchainPendingBlock(
		ctx,
		opClient,
		constant.OP_MAINNET,
		&setup.OpPublicKey,
		&opBlockRetryCount,
		constant.OP_PENDING_BLOCK,
		constant.OP_PENDING_TRANSACTION)
}
