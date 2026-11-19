package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
)

func SweepAvaxBlockchain(ctx context.Context) {
	core.RunEVMSweep(ctx, constant.AVAX_MAINNET)
}
