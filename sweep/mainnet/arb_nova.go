package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	arbitrumNovaBlockRetryCount = make(map[int64]int)
	arbitrumNovaTxRetryCount    = make(map[string]int)

	arbitrumNovaClient NODE_Client.Client
)

func SweepArbitrumNovaBlockchain(ctx context.Context) {
	initArbitrumNova(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepArbitrumNovaBlockchainTransaction(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepArbitrumNovaBlockchainTransactionDetails(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepArbitrumNovaBlockchainPendingBlock(ctx)
			}
		}
	}()
}

func initArbitrumNova(ctx context.Context) {
	core.SetupLatestBlockHeight(ctx, arbitrumNovaClient, constant.ARBITRUM_NOVA)

	setup.SetupCacheBlockHeight(ctx, constant.ARBITRUM_NOVA)

	setup.SetupSweepBlockHeight(ctx, constant.ARBITRUM_NOVA)
}

func SweepArbitrumNovaBlockchainTransaction(ctx context.Context) {
	core.SweepBlockchainTransaction(
		ctx,
		arbitrumNovaClient,
		constant.ARBITRUM_NOVA,
		&setup.ArbitrumNovaPublicKey,
		&setup.ArbitrumNovaSweepBlockHeight,
		&setup.ArbitrumNovaCacheBlockHeight,
		constant.ARBITRUM_NOVA_SWEEP_BLOCK,
		constant.ARBITRUM_NOVA_PENDING_BLOCK,
		constant.ARBITRUM_NOVA_PENDING_TRANSACTION)
}

func SweepArbitrumNovaBlockchainTransactionDetails(ctx context.Context) {
	core.SweepBlockchainTransactionDetails(
		ctx,
		arbitrumNovaClient,
		constant.ARBITRUM_NOVA,
		&setup.ArbitrumNovaPublicKey,
		&arbitrumNovaTxRetryCount,
		constant.ARBITRUM_NOVA_PENDING_TRANSACTION)
}

func SweepArbitrumNovaBlockchainPendingBlock(ctx context.Context) {
	core.SweepBlockchainPendingBlock(
		ctx,
		arbitrumNovaClient,
		constant.ARBITRUM_NOVA,
		&setup.ArbitrumNovaPublicKey,
		&arbitrumNovaBlockRetryCount,
		constant.ARBITRUM_NOVA_PENDING_BLOCK,
		constant.ARBITRUM_NOVA_PENDING_TRANSACTION)
}
