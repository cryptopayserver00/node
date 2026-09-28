package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	btcTestnetBlockRetryCount = make(map[int64]int)
	btcTestnetTxRetryCount    = make(map[string]int)

	btcTestnetClient NODE_Client.Client
)

func SweepBtcTestnetBlockchain(ctx context.Context) {
	initBtcTestnet(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBtcTestnetBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBtcTestnetBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBtcTestnetBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initBtcTestnet(ctx context.Context) {
	core.SetupBtcLatestBlockHeight(ctx, btcTestnetClient, constant.BTC_TESTNET)

	setup.SetupCacheBlockHeight(ctx, constant.BTC_TESTNET)

	setup.SetupSweepBlockHeight(ctx, constant.BTC_TESTNET)
}

func SweepBtcTestnetBlockchainTransaction(ctx context.Context) {
	core.SweepBtcBlockchainTransaction(
		ctx,
		btcTestnetClient,
		constant.BTC_TESTNET,
		&setup.BtcTestnetPublicKey,
		&setup.BtcTestnetSweepBlockHeight,
		&setup.BtcTestnetCacheBlockHeight,
		constant.BTC_TESTNET_SWEEP_BLOCK,
		constant.BTC_TESTNET_PENDING_BLOCK,
		constant.BTC_TESTNET_PENDING_TRANSACTION)
}

func SweepBtcTestnetBlockchainTransactionDetails(ctx context.Context) {
	core.SweepBtcBlockchainTransactionDetails(
		ctx,
		btcTestnetClient,
		constant.BTC_TESTNET,
		&setup.BtcTestnetPublicKey,
		&btcTestnetTxRetryCount,
		constant.BTC_TESTNET_PENDING_TRANSACTION)
}

func SweepBtcTestnetBlockchainPendingBlock(ctx context.Context) {
	core.SweepBtcBlockchainPendingBlock(
		ctx,
		btcTestnetClient,
		constant.BTC_TESTNET,
		&setup.BtcTestnetPublicKey,
		&btcTestnetBlockRetryCount,
		constant.BTC_TESTNET_PENDING_BLOCK,
		constant.BTC_TESTNET_PENDING_TRANSACTION)
}
