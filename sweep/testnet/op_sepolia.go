package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	opSepoliaBlockRetryCount = make(map[int64]int)
	opSepoliaTxRetryCount    = make(map[string]int)

	opSepoliaClient NODE_Client.Client
)

func SweepOpSepoliaBlockchain(ctx context.Context) {

	initOpSepolia(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepOpSepoliaBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepOpSepoliaBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepOpSepoliaBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initOpSepolia(ctx context.Context) {
	core.SetupLatestBlockHeight(ctx, opSepoliaClient, constant.OP_SEPOLIA)

	setup.SetupCacheBlockHeight(ctx, constant.OP_SEPOLIA)

	setup.SetupSweepBlockHeight(ctx, constant.OP_SEPOLIA)
}

func SweepOpSepoliaBlockchainTransaction(ctx context.Context) {
	core.SweepBlockchainTransaction(
		ctx,
		opSepoliaClient,
		constant.OP_SEPOLIA,
		&setup.OpSepoliaPublicKey,
		&setup.OpSepoliaSweepBlockHeight,
		&setup.OpSepoliaCacheBlockHeight,
		constant.OP_SEPOLIA_SWEEP_BLOCK,
		constant.OP_SEPOLIA_PENDING_BLOCK,
		constant.OP_SEPOLIA_PENDING_TRANSACTION)
}

func SweepOpSepoliaBlockchainTransactionDetails(ctx context.Context) {
	core.SweepBlockchainTransactionDetails(
		ctx,
		opSepoliaClient,
		constant.OP_SEPOLIA,
		&setup.OpSepoliaPublicKey,
		&opSepoliaTxRetryCount,
		constant.OP_SEPOLIA_PENDING_TRANSACTION)
}

func SweepOpSepoliaBlockchainPendingBlock(ctx context.Context) {
	core.SweepBlockchainPendingBlock(
		ctx,
		opSepoliaClient,
		constant.OP_SEPOLIA,
		&setup.OpSepoliaPublicKey,
		&opSepoliaBlockRetryCount,
		constant.OP_SEPOLIA_PENDING_BLOCK,
		constant.OP_SEPOLIA_PENDING_TRANSACTION)
}
