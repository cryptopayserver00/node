package service

import (
	"node/model/node/request"
	"node/model/node/response"
)

func (n *NService) GetTonTransactions(req request.GetTonTransactions) ([]response.ClientTransaction, error) {
	return nil, nil
}

func (n *NService) GetTonCoinTransactions(req request.GetTonCoinTransactions) ([]response.ClientTransaction, error) {
	return nil, nil
}

func (n *NService) GetTon20Transactions(req request.GetTon20Transactions) ([]response.ClientTransaction, error) {
	return nil, nil
}
