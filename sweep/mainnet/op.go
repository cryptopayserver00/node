package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
)

func SweepOpBlockchain(ctx context.Context) {
	core.RunEVMSweep(ctx, constant.OP_MAINNET)
}
