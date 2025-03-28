package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

var (
	avaxTestnetSweepCount = make(map[int64]int)

	avaxTestnetClient NODE_Client.Client
)

func SweepAvaxTestnetBlockchain() {
	initAvaxTestnet()

	go func() {
		for {
			SweepAvaxTestnetBlockchainTransaction()
		}
	}()

	go func() {
		for {
			SweepAvaxTestnetBlockchainTransactionDetails()
		}
	}()

	go func() {
		for {
			SweepAvaxTestnetBlockchainPendingBlock()
		}
	}()
}

func initAvaxTestnet() {
	core.SetupLatestBlockHeight(avaxTestnetClient, constant.AVAX_TESTNET)

	setup.SetupCacheBlockHeight(context.Background(), constant.AVAX_TESTNET)

	setup.SetupSweepBlockHeight(context.Background(), constant.AVAX_TESTNET)
}

func SweepAvaxTestnetBlockchainTransaction() {

	core.SweepBlockchainTransaction(
		avaxTestnetClient,
		constant.AVAX_TESTNET,
		&setup.AvaxTestnetPublicKey,
		&avaxTestnetSweepCount,
		&setup.AvaxTestnetSweepBlockHeight,
		&setup.AvaxTestnetCacheBlockHeight,
		constant.AVAX_TESTNET_SWEEP_BLOCK,
		constant.AVAX_TESTNET_PENDING_BLOCK,
		constant.AVAX_TESTNET_PENDING_TRANSACTION)
}

func SweepAvaxTestnetBlockchainTransactionDetails() {

	core.SweepBlockchainTransactionDetails(
		avaxTestnetClient,
		constant.AVAX_TESTNET,
		&setup.AvaxTestnetPublicKey,
		constant.AVAX_TESTNET_PENDING_TRANSACTION)
}

func SweepAvaxTestnetBlockchainPendingBlock() {
	core.SweepBlockchainPendingBlock(
		avaxTestnetClient,
		constant.AVAX_TESTNET,
		&setup.AvaxTestnetPublicKey,
		&avaxTestnetSweepCount,
		constant.AVAX_TESTNET_PENDING_BLOCK,
		constant.AVAX_TESTNET_PENDING_TRANSACTION)
}
