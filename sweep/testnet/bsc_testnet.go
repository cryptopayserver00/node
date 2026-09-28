package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	bscTestnetBlockRetryCount = make(map[int64]int)
	bscTestnetTxRetryCount    = make(map[string]int)

	bscTestnetClient NODE_Client.Client
)

func SweepBscTestnetBlockchain(ctx context.Context) {
	initBscTestnet(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBscTestnetBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBscTestnetBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBscTestnetBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initBscTestnet(ctx context.Context) {
	core.SetupLatestBlockHeight(ctx, bscTestnetClient, constant.BSC_TESTNET)

	setup.SetupCacheBlockHeight(ctx, constant.BSC_TESTNET)

	setup.SetupSweepBlockHeight(ctx, constant.BSC_TESTNET)
}

func SweepBscTestnetBlockchainTransaction(ctx context.Context) {
	core.SweepBlockchainTransaction(
		ctx,
		bscTestnetClient,
		constant.BSC_TESTNET,
		&setup.BscTestnetPublicKey,
		&setup.BscTestnetSweepBlockHeight,
		&setup.BscTestnetCacheBlockHeight,
		constant.BSC_TESTNET_SWEEP_BLOCK,
		constant.BSC_TESTNET_PENDING_BLOCK,
		constant.BSC_TESTNET_PENDING_TRANSACTION)
}

func SweepBscTestnetBlockchainTransactionDetails(ctx context.Context) {
	core.SweepBlockchainTransactionDetails(
		ctx,
		bscTestnetClient,
		constant.BSC_TESTNET,
		&setup.BscTestnetPublicKey,
		&bscTestnetTxRetryCount,
		constant.BSC_TESTNET_PENDING_TRANSACTION)
}

func SweepBscTestnetBlockchainPendingBlock(ctx context.Context) {
	core.SweepBlockchainPendingBlock(
		ctx,
		bscTestnetClient,
		constant.BSC_TESTNET,
		&setup.BscTestnetPublicKey,
		&bscTestnetBlockRetryCount,
		constant.BSC_TESTNET_PENDING_BLOCK,
		constant.BSC_TESTNET_PENDING_TRANSACTION)
}
