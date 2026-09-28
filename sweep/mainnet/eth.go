package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	ethBlockRetryCount = make(map[int64]int)
	ethTxRetryCount    = make(map[string]int)

	ethClient NODE_Client.Client
)

func SweepEthBlockchain(ctx context.Context) {

	initEth(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepEthBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepEthBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepEthBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initEth(ctx context.Context) {
	core.SetupLatestBlockHeight(ctx, ethClient, constant.ETH_MAINNET)

	setup.SetupCacheBlockHeight(ctx, constant.ETH_MAINNET)

	setup.SetupSweepBlockHeight(ctx, constant.ETH_MAINNET)
}

func SweepEthBlockchainTransaction(ctx context.Context) {
	core.SweepBlockchainTransaction(
		ctx,
		ethClient,
		constant.ETH_MAINNET,
		&setup.EthPublicKey,
		&setup.EthSweepBlockHeight,
		&setup.EthCacheBlockHeight,
		constant.ETH_SWEEP_BLOCK,
		constant.ETH_PENDING_BLOCK,
		constant.ETH_PENDING_TRANSACTION)
}

func SweepEthBlockchainTransactionDetails(ctx context.Context) {
	core.SweepBlockchainTransactionDetails(
		ctx,
		ethClient,
		constant.ETH_MAINNET,
		&setup.EthPublicKey,
		&ethTxRetryCount,
		constant.ETH_PENDING_TRANSACTION)
}

func SweepEthBlockchainPendingBlock(ctx context.Context) {
	core.SweepBlockchainPendingBlock(
		ctx,
		ethClient,
		constant.ETH_MAINNET,
		&setup.EthPublicKey,
		&ethBlockRetryCount,
		constant.ETH_PENDING_BLOCK,
		constant.ETH_PENDING_TRANSACTION)
}
