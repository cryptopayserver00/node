package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
)

func SweepEthSepoliaBlockchain(ctx context.Context) {
	core.RunEVMSweep(ctx, constant.ETH_SEPOLIA)
}
