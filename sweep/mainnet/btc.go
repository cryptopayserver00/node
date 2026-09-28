package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	btcBlockRetryCount = make(map[int64]int)
	btcTxRetryCount    = make(map[string]int)

	btcClient NODE_Client.Client
)

func SweepBtcBlockchain(ctx context.Context) {
	initBtc(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBtcBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBtcBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBtcBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initBtc(ctx context.Context) {
	core.SetupBtcLatestBlockHeight(ctx, btcClient, constant.BTC_MAINNET)

	setup.SetupCacheBlockHeight(ctx, constant.BTC_MAINNET)

	setup.SetupSweepBlockHeight(ctx, constant.BTC_MAINNET)
}

func SweepBtcBlockchainTransaction(ctx context.Context) {
	core.SweepBtcBlockchainTransaction(
		ctx,
		btcClient,
		constant.BTC_MAINNET,
		&setup.BtcPublicKey,
		&setup.BtcSweepBlockHeight,
		&setup.BtcCacheBlockHeight,
		constant.BTC_SWEEP_BLOCK,
		constant.BTC_PENDING_BLOCK,
		constant.BTC_PENDING_TRANSACTION)
}

func SweepBtcBlockchainTransactionDetails(ctx context.Context) {
	core.SweepBtcBlockchainTransactionDetails(
		ctx,
		btcClient,
		constant.BTC_MAINNET,
		&setup.BtcPublicKey,
		&btcTxRetryCount,
		constant.BTC_PENDING_TRANSACTION)
}

func SweepBtcBlockchainPendingBlock(ctx context.Context) {
	core.SweepBtcBlockchainPendingBlock(
		ctx,
		btcClient,
		constant.BTC_MAINNET,
		&setup.BtcPublicKey,
		&btcBlockRetryCount,
		constant.BTC_PENDING_BLOCK,
		constant.BTC_PENDING_TRANSACTION)
}
