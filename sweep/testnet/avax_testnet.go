package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	avaxTestnetBlockRetryCount = make(map[int64]int)
	avaxTestnetTxRetryCount    = make(map[string]int)

	avaxTestnetClient NODE_Client.Client
)

func SweepAvaxTestnetBlockchain(ctx context.Context) {
	initAvaxTestnet(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepAvaxTestnetBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepAvaxTestnetBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepAvaxTestnetBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initAvaxTestnet(ctx context.Context) {
	core.SetupLatestBlockHeight(ctx, avaxTestnetClient, constant.AVAX_TESTNET)

	setup.SetupCacheBlockHeight(ctx, constant.AVAX_TESTNET)

	setup.SetupSweepBlockHeight(ctx, constant.AVAX_TESTNET)
}

func SweepAvaxTestnetBlockchainTransaction(ctx context.Context) {
	core.SweepBlockchainTransaction(
		ctx,
		avaxTestnetClient,
		constant.AVAX_TESTNET,
		&setup.AvaxTestnetPublicKey,
		&setup.AvaxTestnetSweepBlockHeight,
		&setup.AvaxTestnetCacheBlockHeight,
		constant.AVAX_TESTNET_SWEEP_BLOCK,
		constant.AVAX_TESTNET_PENDING_BLOCK,
		constant.AVAX_TESTNET_PENDING_TRANSACTION)
}

func SweepAvaxTestnetBlockchainTransactionDetails(ctx context.Context) {
	core.SweepBlockchainTransactionDetails(
		ctx,
		avaxTestnetClient,
		constant.AVAX_TESTNET,
		&setup.AvaxTestnetPublicKey,
		&avaxTestnetTxRetryCount,
		constant.AVAX_TESTNET_PENDING_TRANSACTION)
}

func SweepAvaxTestnetBlockchainPendingBlock(ctx context.Context) {
	core.SweepBlockchainPendingBlock(
		ctx,
		avaxTestnetClient,
		constant.AVAX_TESTNET,
		&setup.AvaxTestnetPublicKey,
		&avaxTestnetBlockRetryCount,
		constant.AVAX_TESTNET_PENDING_BLOCK,
		constant.AVAX_TESTNET_PENDING_TRANSACTION)
}
