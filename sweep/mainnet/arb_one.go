package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	arbitrumOneSweepCount = make(map[int64]int)

	arbitrumOneClient NODE_Client.Client
)

func SweepArbitrumOneBlockchain() {

	initArbitrumOne()

	go func() {
		for {
			SweepArbitrumOneBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepArbitrumOneBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepArbitrumOneBlockchainPendingBlock()
		}
	}()
}

func initArbitrumOne() {
	core.SetupLatestBlockHeight(arbitrumOneClient, constant.ARBITRUM_ONE)

	setup.SetupCacheBlockHeight(context.Background(), constant.ARBITRUM_ONE)

	setup.SetupSweepBlockHeight(context.Background(), constant.ARBITRUM_ONE)
}

func SweepArbitrumOneBlockchainTransaction() {
	core.SweepBlockchainTransaction(
		arbitrumOneClient,
		constant.ARBITRUM_ONE,
		&setup.ArbitrumOnePublicKey,
		&arbitrumOneSweepCount,
		&setup.ArbitrumOneSweepBlockHeight,
		&setup.ArbitrumOneCacheBlockHeight,
		constant.ARBITRUM_ONE_SWEEP_BLOCK,
		constant.ARBITRUM_ONE_PENDING_BLOCK,
		constant.ARBITRUM_ONE_PENDING_TRANSACTION)
}

func SweepArbitrumOneBlockchainTransactionDetails() {
	core.SweepBlockchainTransactionDetails(
		arbitrumOneClient,
		constant.ARBITRUM_ONE,
		&setup.ArbitrumOnePublicKey,
		constant.ARBITRUM_ONE_PENDING_TRANSACTION)
}

func SweepArbitrumOneBlockchainPendingBlock() {
	core.SweepBlockchainPendingBlock(
		arbitrumOneClient,
		constant.ARBITRUM_ONE,
		&setup.ArbitrumOnePublicKey,
		&arbitrumOneSweepCount,
		constant.ARBITRUM_ONE_PENDING_BLOCK,
		constant.ARBITRUM_ONE_PENDING_TRANSACTION)
}
