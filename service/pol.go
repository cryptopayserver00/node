package service

import (
	"context"
	"node/model/node/request"
	"node/model/node/response"
)

func (n *NService) GetPolTransactions(ctx context.Context, req request.GetPolTransactions) ([]response.ClientTransaction, error) {
	return n.GetEvmTransactions(ctx, req)
}
