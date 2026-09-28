package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	ltcBlockRetryCount = make(map[int64]int)
	ltcTxRetryCount    = make(map[string]int)

	ltcClient NODE_Client.Client
)

func SweepLtcBlockchain(ctx context.Context) {
	initLtc(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepLtcBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepLtcBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepLtcBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initLtc(ctx context.Context) {
	core.SetupLtcLatestBlockHeight(ctx, ltcClient, constant.LTC_MAINNET)

	setup.SetupCacheBlockHeight(ctx, constant.LTC_MAINNET)

	setup.SetupSweepBlockHeight(ctx, constant.LTC_MAINNET)
}

func SweepLtcBlockchainTransaction(ctx context.Context) {
	core.SweepLtcBlockchainTransaction(
		ctx,
		ltcClient,
		constant.LTC_MAINNET,
		&setup.LtcPublicKey,
		&setup.LtcSweepBlockHeight,
		&setup.LtcCacheBlockHeight,
		constant.LTC_SWEEP_BLOCK,
		constant.LTC_PENDING_BLOCK,
		constant.LTC_PENDING_TRANSACTION)
}

func SweepLtcBlockchainTransactionDetails(ctx context.Context) {
	core.SweepLtcBlockchainTransactionDetails(
		ctx,
		ltcClient,
		constant.LTC_MAINNET,
		&setup.LtcPublicKey,
		&ltcTxRetryCount,
		constant.LTC_PENDING_TRANSACTION)
}

func SweepLtcBlockchainPendingBlock(ctx context.Context) {
	core.SweepLtcBlockchainPendingBlock(
		ctx,
		ltcClient,
		constant.LTC_MAINNET,
		&setup.LtcPublicKey,
		&ltcBlockRetryCount,
		constant.LTC_PENDING_BLOCK,
		constant.LTC_PENDING_TRANSACTION)
}
