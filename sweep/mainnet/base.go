package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
)

func SweepBaseBlockchain(ctx context.Context) {
	core.RunEVMSweep(ctx, constant.BASE_MAINNET)
}
