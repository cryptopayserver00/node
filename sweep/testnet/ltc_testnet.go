package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	ltcTestnetBlockRetryCount = make(map[int64]int)
	ltcTestnetTxRetryCount    = make(map[string]int)

	ltcTestnetClient NODE_Client.Client
)

func SweepLtcTestnetBlockchain(ctx context.Context) {
	initLtcTestnet(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepLtcTestnetBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepLtcTestnetBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepLtcTestnetBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initLtcTestnet(ctx context.Context) {
	core.SetupLtcLatestBlockHeight(ctx, ltcTestnetClient, constant.LTC_TESTNET)

	setup.SetupCacheBlockHeight(ctx, constant.LTC_TESTNET)

	setup.SetupSweepBlockHeight(ctx, constant.LTC_TESTNET)
}

func SweepLtcTestnetBlockchainTransaction(ctx context.Context) {
	core.SweepLtcBlockchainTransaction(
		ctx,
		ltcTestnetClient,
		constant.LTC_TESTNET,
		&setup.LtcTestnetPublicKey,
		&setup.LtcTestnetSweepBlockHeight,
		&setup.LtcTestnetCacheBlockHeight,
		constant.LTC_TESTNET_SWEEP_BLOCK,
		constant.LTC_TESTNET_PENDING_BLOCK,
		constant.LTC_TESTNET_PENDING_TRANSACTION)
}

func SweepLtcTestnetBlockchainTransactionDetails(ctx context.Context) {
	core.SweepLtcBlockchainTransactionDetails(
		ctx,
		ltcTestnetClient,
		constant.LTC_TESTNET,
		&setup.LtcTestnetPublicKey,
		&ltcTestnetTxRetryCount,
		constant.LTC_TESTNET_PENDING_TRANSACTION)
}

func SweepLtcTestnetBlockchainPendingBlock(ctx context.Context) {
	core.SweepLtcBlockchainPendingBlock(
		ctx,
		ltcTestnetClient,
		constant.LTC_TESTNET,
		&setup.LtcTestnetPublicKey,
		&ltcTestnetBlockRetryCount,
		constant.LTC_TESTNET_PENDING_BLOCK,
		constant.LTC_TESTNET_PENDING_TRANSACTION)
}
