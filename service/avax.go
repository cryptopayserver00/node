package service

import (
	"context"
	"fmt"
	"node/global"
	"node/model/node/request"
	"node/model/node/response"
)

func (n *NService) GetAvaxTransactions(ctx context.Context, req request.GetAvaxTransactions) ([]response.ClientTransaction, error) {
	txs, err := n.GetEvmTransactionsByExplorer(ctx, req.ChainId, req.Address, req.Asset)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("avax explorer txs failed: %v", err))
		return []response.ClientTransaction{}, nil
	}
	return txs, nil
}
