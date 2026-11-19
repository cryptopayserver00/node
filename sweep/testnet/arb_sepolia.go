package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
)

func SweepArbitrumSepoliaBlockchain(ctx context.Context) {
	core.RunEVMSweep(ctx, constant.ARBITRUM_SEPOLIA)
}
