package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
)

func SweepBscTestnetBlockchain(ctx context.Context) {
	core.RunEVMSweep(ctx, constant.BSC_TESTNET)
}
