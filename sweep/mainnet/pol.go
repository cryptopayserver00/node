package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
)

func SweepPolBlockchain(ctx context.Context) {
	core.RunEVMSweep(ctx, constant.POL_MAINNET)
}
