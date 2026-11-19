package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
)

func SweepBaseSepoliaBlockchain(ctx context.Context) {
	core.RunEVMSweep(ctx, constant.BASE_SEPOLIA)
}
