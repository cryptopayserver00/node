package adapter

import (
	"context"
	"node/global/constant"
	"node/sweep/core"
)

type evmAdapter struct {
	chainID uint
}

func NewEVMAdapter(chainID uint) Adapter {
	return &evmAdapter{chainID: chainID}
}

func (a *evmAdapter) ChainID() uint { return a.chainID }

func (a *evmAdapter) Name() string {
	return constant.GetChainName(a.chainID)
}

func (a *evmAdapter) Run(ctx context.Context) {
	core.RunEVMSweep(ctx, a.chainID)
}
