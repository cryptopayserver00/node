package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	polTestnetBlockRetryCount = make(map[int64]int)
	polTestnetTxRetryCount    = make(map[string]int)

	polTestnetClient NODE_Client.Client
)

func SweepPolTestnetBlockchain(ctx context.Context) {
	initPolTestnet(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepPolTestnetBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepPolTestnetBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepPolTestnetBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initPolTestnet(ctx context.Context) {
	core.SetupLatestBlockHeight(ctx, polTestnetClient, constant.POL_TESTNET)

	setup.SetupCacheBlockHeight(ctx, constant.POL_TESTNET)

	setup.SetupSweepBlockHeight(ctx, constant.POL_TESTNET)
}

func SweepPolTestnetBlockchainTransaction(ctx context.Context) {

	core.SweepBlockchainTransaction(
		ctx,
		polTestnetClient,
		constant.POL_TESTNET,
		&setup.PolTestnetPublicKey,
		&setup.PolTestnetSweepBlockHeight,
		&setup.PolTestnetCacheBlockHeight,
		constant.POL_TESTNET_SWEEP_BLOCK,
		constant.POL_TESTNET_PENDING_BLOCK,
		constant.POL_TESTNET_PENDING_TRANSACTION)
}

func SweepPolTestnetBlockchainTransactionDetails(ctx context.Context) {

	core.SweepBlockchainTransactionDetails(
		ctx,
		polTestnetClient,
		constant.POL_TESTNET,
		&setup.PolTestnetPublicKey,
		&polTestnetTxRetryCount,
		constant.POL_TESTNET_PENDING_TRANSACTION)
}

func SweepPolTestnetBlockchainPendingBlock(ctx context.Context) {
	core.SweepBlockchainPendingBlock(
		ctx,
		polTestnetClient,
		constant.POL_TESTNET,
		&setup.PolTestnetPublicKey,
		&polTestnetBlockRetryCount,
		constant.POL_TESTNET_PENDING_BLOCK,
		constant.POL_TESTNET_PENDING_TRANSACTION)
}
