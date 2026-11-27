package service

import (
	"context"
	"node/model/node/request"
	"node/model/node/response"
)

func (n *NService) GetBaseTransactions(ctx context.Context, req request.GetBaseTransactions) ([]response.ClientTransaction, error) {
	return n.GetEvmTransactions(ctx, req)
}
