package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	ethGoerliSweepCount = make(map[int64]int)

	ethGoerliClient NODE_Client.Client
)

func SweepEthGoerliBlockchain() {
	initEthGoerli()

	go func() {
		for {
			SweepEthGoerliBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepEthGoerliBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepEthGoerliBlockchainPendingBlock()
		}
	}()
}

func initEthGoerli() {
	core.SetupLatestBlockHeight(ethGoerliClient, constant.ETH_GOERLI)

	setup.SetupCacheBlockHeight(context.Background(), constant.ETH_GOERLI)

	setup.SetupSweepBlockHeight(context.Background(), constant.ETH_GOERLI)
}

func SweepEthGoerliBlockchainTransaction() {
	core.SweepBlockchainTransaction(
		ethGoerliClient,
		constant.ETH_GOERLI,
		&setup.EthGoerliPublicKey,
		&ethGoerliSweepCount,
		&setup.EthGoerliSweepBlockHeight,
		&setup.EthGoerliCacheBlockHeight,
		constant.ETH_GOERLI_SWEEP_BLOCK,
		constant.ETH_GOERLI_PENDING_BLOCK,
		constant.ETH_GOERLI_PENDING_TRANSACTION)
}

func SweepEthGoerliBlockchainTransactionDetails() {
	core.SweepBlockchainTransactionDetails(
		ethGoerliClient,
		constant.ETH_GOERLI,
		&setup.EthGoerliPublicKey,
		constant.ETH_GOERLI_PENDING_TRANSACTION)
}

func SweepEthGoerliBlockchainPendingBlock() {
	core.SweepBlockchainPendingBlock(
		ethGoerliClient,
		constant.ETH_GOERLI,
		&setup.EthGoerliPublicKey,
		constant.ETH_GOERLI_PENDING_BLOCK,
		constant.ETH_GOERLI_PENDING_TRANSACTION)
}
