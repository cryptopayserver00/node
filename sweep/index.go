package sweep

import (
	"context"
	"fmt"
	"node/global"
	"node/sweep/adapter"
	"node/sweep/setup"

	"go.uber.org/zap"
)

func RunBlockSweep(ctx context.Context) {
	setup.SetupPublicKey(ctx)

	adapters := adapter.EnabledAdapters()
	if len(adapters) == 0 {
		global.NODE_LOG.Warn("sweep: no chain adapter enabled")
		return
	}

	for _, a := range adapters {
		global.NODE_LOG.Info("sweep: start adapter",
			zap.Uint("chain_id", a.ChainID()),
			zap.String("name", a.Name()),
		)
	}
	global.NODE_LOG.Info(fmt.Sprintf("sweep: launching %d adapters", len(adapters)))
	adapter.Runner(ctx, adapters)
}
