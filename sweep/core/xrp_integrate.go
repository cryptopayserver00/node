package core

import (
	"node/utils"
	NODE_Client "node/utils/http"
)

func SetupXrpLatestBlockHeight(client NODE_Client.Client, chainId uint) {
}

func SweepXrpBlockchainTransaction(
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	sweepCount *map[int64]int,
	sweepBlockHeight, cacheBlockHeight *int64,
	constantSweepBlock, constantPendingBlock, constantPendingTransaction string) {
	defer utils.HandlePanic()
}

func SweepXrpBlockchainTransactionDetails(
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	constantPendingTransaction string,
) {
	defer utils.HandlePanic()

}

func SweepXrpBlockchainPendingBlock(
	client NODE_Client.Client,
	chainId uint,
	publicKey *[]string,
	constantPendingBlock, constantPendingTransaction string,
) {
	defer utils.HandlePanic()
}
