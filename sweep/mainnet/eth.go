package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
)

func SweepEthBlockchain(ctx context.Context) {
	core.RunEVMSweep(ctx, constant.ETH_MAINNET)
}
