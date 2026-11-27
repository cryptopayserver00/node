package service

import (
	"context"
	"node/model/node/request"
	"node/model/node/response"
)

func (n *NService) GetAvaxTransactions(ctx context.Context, req request.GetAvaxTransactions) ([]response.ClientTransaction, error) {
	return n.GetEvmTransactions(ctx, req)
}
