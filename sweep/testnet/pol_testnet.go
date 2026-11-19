package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
)

func SweepPolTestnetBlockchain(ctx context.Context) {
	core.RunEVMSweep(ctx, constant.POL_TESTNET)
}
