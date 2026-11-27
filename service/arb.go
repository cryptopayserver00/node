package service

import (
	"context"
	"node/model/node/request"
	"node/model/node/response"
)

func (n *NService) GetArbTransactions(ctx context.Context, req request.GetArbTransactions) ([]response.ClientTransaction, error) {
	return n.GetEvmTransactions(ctx, req)
}
