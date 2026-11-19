package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
)

func SweepAvaxTestnetBlockchain(ctx context.Context) {
	core.RunEVMSweep(ctx, constant.AVAX_TESTNET)
}
