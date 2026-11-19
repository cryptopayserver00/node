package core

import (
	"context"
	"node/global"
	"node/global/constant"
	"node/sweep/setup"
	NODE_Client "node/utils/http"
)

// RunEVMSweep 启动通用 EVM 扫块三循环（扫块 / 解析 pending tx / 重试失败块）。
func RunEVMSweep(ctx context.Context, chainId uint) {
	st, ok := setup.GetChainState(chainId)
	if !ok {
		global.NODE_LOG.Error("RunEVMSweep: unsupported chain")
		return
	}

	client := NODE_Client.Client{
		OnSuccess: func(url string) { constant.ReportRPCSuccess(chainId, url) },
		OnFailure: func(url string) { constant.ReportRPCFailure(chainId, url) },
	}
	blockRetryCount := make(map[int64]int)
	txRetryCount := make(map[string]int)

	SetupLatestBlockHeight(ctx, client, chainId)
	setup.SetupCacheBlockHeight(ctx, chainId)
	setup.SetupSweepBlockHeight(ctx, chainId)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBlockchainTransaction(
					ctx,
					client,
					chainId,
					st.PublicKeys,
					st.Sweep,
					st.Cache,
					st.SweepKey,
					st.PendingBlockKey,
					st.PendingTransactionKey,
				)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBlockchainTransactionDetails(
					ctx,
					client,
					chainId,
					st.PublicKeys,
					&txRetryCount,
					st.PendingTransactionKey,
				)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				SweepBlockchainPendingBlock(
					ctx,
					client,
					chainId,
					st.PublicKeys,
					&blockRetryCount,
					st.PendingBlockKey,
					st.PendingTransactionKey,
				)
			}
		}
	}()
}
