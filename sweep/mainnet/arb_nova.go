package mainnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
)

func SweepArbitrumNovaBlockchain(ctx context.Context) {
	core.RunEVMSweep(ctx, constant.ARBITRUM_NOVA)
}
