package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	avaxBlockRetryCount = make(map[int64]int)
	avaxTxRetryCount    = make(map[string]int)

	avaxClient NODE_Client.Client
)

func SweepAvaxBlockchain(ctx context.Context) {
	initAvax(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepAvaxBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepAvaxBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepAvaxBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initAvax(ctx context.Context) {
	core.SetupLatestBlockHeight(ctx, avaxClient, constant.AVAX_MAINNET)

	setup.SetupCacheBlockHeight(ctx, constant.AVAX_MAINNET)

	setup.SetupSweepBlockHeight(ctx, constant.AVAX_MAINNET)
}

func SweepAvaxBlockchainTransaction(ctx context.Context) {
	core.SweepBlockchainTransaction(
		ctx,
		avaxClient,
		constant.AVAX_MAINNET,
		&setup.AvaxPublicKey,
		&setup.AvaxSweepBlockHeight,
		&setup.AvaxCacheBlockHeight,
		constant.AVAX_SWEEP_BLOCK,
		constant.AVAX_PENDING_BLOCK,
		constant.AVAX_PENDING_TRANSACTION)
}

func SweepAvaxBlockchainTransactionDetails(ctx context.Context) {
	core.SweepBlockchainTransactionDetails(
		ctx,
		avaxClient,
		constant.AVAX_MAINNET,
		&setup.AvaxPublicKey,
		&avaxTxRetryCount,
		constant.AVAX_PENDING_TRANSACTION)
}

func SweepAvaxBlockchainPendingBlock(ctx context.Context) {
	core.SweepBlockchainPendingBlock(
		ctx,
		avaxClient,
		constant.AVAX_MAINNET,
		&setup.AvaxPublicKey,
		&avaxBlockRetryCount,
		constant.AVAX_PENDING_BLOCK,
		constant.AVAX_PENDING_TRANSACTION)
}
