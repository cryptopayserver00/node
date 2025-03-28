package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	avaxSweepCount = make(map[int64]int)

	avaxClient NODE_Client.Client
)

func SweepAvaxBlockchain() {
	initAvax()

	go func() {
		for {
			SweepAvaxBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepAvaxBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepAvaxBlockchainPendingBlock()
		}
	}()
}

func initAvax() {
	core.SetupLatestBlockHeight(avaxClient, constant.AVAX_MAINNET)

	setup.SetupCacheBlockHeight(context.Background(), constant.AVAX_MAINNET)

	setup.SetupSweepBlockHeight(context.Background(), constant.AVAX_MAINNET)
}

func SweepAvaxBlockchainTransaction() {

	core.SweepBlockchainTransaction(
		avaxClient,
		constant.AVAX_MAINNET,
		&setup.AvaxPublicKey,
		&avaxSweepCount,
		&setup.AvaxSweepBlockHeight,
		&setup.AvaxCacheBlockHeight,
		constant.AVAX_SWEEP_BLOCK,
		constant.AVAX_PENDING_BLOCK,
		constant.AVAX_PENDING_TRANSACTION)
}

func SweepAvaxBlockchainTransactionDetails() {

	core.SweepBlockchainTransactionDetails(
		avaxClient,
		constant.AVAX_MAINNET,
		&setup.AvaxPublicKey,
		constant.AVAX_PENDING_TRANSACTION)
}

func SweepAvaxBlockchainPendingBlock() {
	core.SweepBlockchainPendingBlock(
		avaxClient,
		constant.AVAX_MAINNET,
		&setup.AvaxPublicKey,
		&avaxSweepCount,
		constant.AVAX_PENDING_BLOCK,
		constant.AVAX_PENDING_TRANSACTION)
}
