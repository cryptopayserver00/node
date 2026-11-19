package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
)

func SweepBscBlockchain(ctx context.Context) {
	core.RunEVMSweep(ctx, constant.BSC_MAINNET)
}
