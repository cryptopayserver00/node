package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	arbitrumGoerliSweepCount = make(map[int64]int)

	arbitrumGoerliClient NODE_Client.Client
)

func SweepArbitrumGoerliBlockchain() {

	initArbitrumGoerli()

	go func() {
		for {
			SweepArbitrumGoerliBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepArbitrumGoerliBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepArbitrumGoerliBlockchainPendingBlock()
		}
	}()
}

func initArbitrumGoerli() {
	core.SetupLatestBlockHeight(arbitrumGoerliClient, constant.ARBITRUM_GOERLI)

	setup.SetupCacheBlockHeight(context.Background(), constant.ARBITRUM_GOERLI)

	setup.SetupSweepBlockHeight(context.Background(), constant.ARBITRUM_GOERLI)
}

func SweepArbitrumGoerliBlockchainTransaction() {
	core.SweepBlockchainTransaction(
		arbitrumGoerliClient,
		constant.ARBITRUM_GOERLI,
		&setup.ArbitrumGoerliPublicKey,
		&arbitrumGoerliSweepCount,
		&setup.ArbitrumGoerliSweepBlockHeight,
		&setup.ArbitrumGoerliCacheBlockHeight,
		constant.ARBITRUM_GOERLI_SWEEP_BLOCK,
		constant.ARBITRUM_GOERLI_PENDING_BLOCK,
		constant.ARBITRUM_GOERLI_PENDING_TRANSACTION)
}

func SweepArbitrumGoerliBlockchainTransactionDetails() {
	core.SweepBlockchainTransactionDetails(
		arbitrumGoerliClient,
		constant.ARBITRUM_GOERLI,
		&setup.ArbitrumGoerliPublicKey,
		constant.ARBITRUM_GOERLI_PENDING_TRANSACTION)
}

func SweepArbitrumGoerliBlockchainPendingBlock() {
	core.SweepBlockchainPendingBlock(
		arbitrumGoerliClient,
		constant.ARBITRUM_GOERLI,
		&setup.ArbitrumGoerliPublicKey,
		constant.ARBITRUM_GOERLI_PENDING_BLOCK,
		constant.ARBITRUM_GOERLI_PENDING_TRANSACTION)
}
