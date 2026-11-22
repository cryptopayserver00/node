package adapter

import (
	"context"
	"node/global/constant"
)

type funcAdapter struct {
	chainID uint
	name    string
	runFn   func(context.Context)
}

func NewFuncAdapter(chainID uint, name string, runFn func(context.Context)) Adapter {
	return &funcAdapter{chainID: chainID, name: name, runFn: runFn}
}

func (a *funcAdapter) ChainID() uint { return a.chainID }

func (a *funcAdapter) Name() string {
	if a.name != "" {
		return a.name
	}
	return constant.GetChainName(a.chainID)
}

func (a *funcAdapter) Run(ctx context.Context) {
	if a.runFn != nil {
		a.runFn(ctx)
	}
}
