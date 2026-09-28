package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	tronBlockRetryCount = make(map[int64]int)
	tronTxRetryCount    = make(map[string]int)

	tronClient NODE_Client.Client
)

func SweepTronBlockchain(ctx context.Context) {
	initTron(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepTronBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepTronBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepTronBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initTron(ctx context.Context) {
	core.SetupTronLatestBlockHeight(ctx, tronClient, constant.TRON_MAINNET)

	setup.SetupCacheBlockHeight(ctx, constant.TRON_MAINNET)

	setup.SetupSweepBlockHeight(ctx, constant.TRON_MAINNET)
}

func SweepTronBlockchainTransaction(ctx context.Context) {
	core.SweepTronBlockchainTransaction(
		ctx,
		tronClient,
		constant.TRON_MAINNET,
		&setup.TronPublicKey,
		&setup.TronSweepBlockHeight,
		&setup.TronCacheBlockHeight,
		constant.TRON_SWEEP_BLOCK,
		constant.TRON_PENDING_BLOCK,
		constant.TRON_PENDING_TRANSACTION)
}

func SweepTronBlockchainTransactionDetails(ctx context.Context) {
	core.SweepTronBlockchainTransactionDetails(
		ctx,
		tronClient,
		constant.TRON_MAINNET,
		&setup.TronPublicKey,
		&tronTxRetryCount,
		constant.TRON_PENDING_TRANSACTION)
}

func SweepTronBlockchainPendingBlock(ctx context.Context) {
	core.SweepTronBlockchainPendingBlock(
		ctx,
		tronClient,
		constant.TRON_MAINNET,
		&setup.TronPublicKey,
		&tronBlockRetryCount,
		constant.TRON_PENDING_BLOCK,
		constant.TRON_PENDING_TRANSACTION)
}
