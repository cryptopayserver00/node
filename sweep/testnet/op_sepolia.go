package testnet

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
)

func SweepOpSepoliaBlockchain(ctx context.Context) {
	core.RunEVMSweep(ctx, constant.OP_SEPOLIA)
}
