package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
)

func SweepArbitrumOneBlockchain(ctx context.Context) {
	core.RunEVMSweep(ctx, constant.ARBITRUM_ONE)
}
