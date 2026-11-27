package service

import (
	"context"
	"node/model/node/request"
	"node/model/node/response"
)

func (n *NService) GetArbNovaTransactions(ctx context.Context, req request.GetArbNovaTransactions) ([]response.ClientTransaction, error) {
	return n.GetEvmTransactions(ctx, req)
}
