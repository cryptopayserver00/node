package service

import (
	"context"
	"fmt"
	"math/big"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
	"node/utils"
	"sort"
	"strings"
	"time"
)

type tonCenterTxResponse struct {
	Ok     bool `json:"ok"`
	Result []struct {
		TransactionID struct {
			Hash string `json:"hash"`
			Lt   string `json:"lt"`
		} `json:"transaction_id"`
		Utime   int64  `json:"utime"`
		Fee     string `json:"fee"`
		InMsg   *tonCenterMsg `json:"in_msg"`
		OutMsgs []tonCenterMsg `json:"out_msgs"`
	} `json:"result"`
	Error string `json:"error"`
}

type tonCenterMsg struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Value       string `json:"value"`
	Message     string `json:"message"`
}

func (n *NService) GetTonTransactions(ctx context.Context, req request.GetTonTransactions) ([]response.ClientTransaction, error) {
	txs, err := n.fetchTonCenterTransactions(ctx, req.ChainId, req.Address)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("ton txs failed: %v", err))
		return []response.ClientTransaction{}, nil
	}
	asset := strings.TrimSpace(req.Asset)
	if asset == "" {
		return txs, nil
	}
	out := make([]response.ClientTransaction, 0, len(txs))
	for _, tx := range txs {
		if strings.EqualFold(tx.Asset, asset) {
			out = append(out, tx)
		}
	}
	return out, nil
}

func (n *NService) GetTonCoinTransactions(ctx context.Context, req request.GetTonCoinTransactions) ([]response.ClientTransaction, error) {
	return n.GetTonTransactions(ctx, request.GetTonTransactions{
		ChainId: req.ChainId,
		Address: req.Address,
		Asset:   constant.TON,
	})
}

func (n *NService) GetTon20Transactions(ctx context.Context, req request.GetTon20Transactions) ([]response.ClientTransaction, error) {
	// Toncenter free API exposes native message value history; jetton history needs a richer indexer.
	// Return empty rather than inventing incorrect jetton transfers.
	_ = ctx
	_ = req
	return []response.ClientTransaction{}, nil
}

func (n *NService) fetchTonCenterTransactions(ctx context.Context, chainId uint, address string) ([]response.ClientTransaction, error) {
	base := constant.GetTonCenterAPIBase(chainId)
	if base == "" {
		return nil, fmt.Errorf("unsupported ton chain %d", chainId)
	}

	httpClient := client
	httpClient.URL = fmt.Sprintf("%s/getTransactions?address=%s&limit=50", base, address)
	httpClient.Timeout = 20 * time.Second

	var resp tonCenterTxResponse
	if err := httpClient.HTTPGet(ctx, &resp); err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("toncenter error: %s", firstNonEmpty(resp.Error, "request failed"))
	}

	var out []response.ClientTransaction
	for _, row := range resp.Result {
		hash := row.TransactionID.Hash
		if hash == "" {
			continue
		}

		// Received via in_msg
		if row.InMsg != nil && strings.TrimSpace(row.InMsg.Value) != "" && row.InMsg.Value != "0" {
			if model, err := decodeTonMsg(chainId, address, hash, int(row.Utime*1000), *row.InMsg); err == nil {
				out = append(out, model)
			}
		}

		// Sent via out_msgs
		for _, msg := range row.OutMsgs {
			if strings.TrimSpace(msg.Value) == "" || msg.Value == "0" {
				continue
			}
			if model, err := decodeTonMsg(chainId, address, hash, int(row.Utime*1000), msg); err == nil {
				out = append(out, model)
			}
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].BlockTimestamp > out[j].BlockTimestamp
	})
	return out, nil
}

func decodeTonMsg(chainId uint, address, hash string, tsMs int, msg tonCenterMsg) (response.ClientTransaction, error) {
	var model response.ClientTransaction
	amountInt, ok := new(big.Int).SetString(msg.Value, 10)
	if !ok || amountInt.Sign() <= 0 {
		return model, fmt.Errorf("skip empty ton value")
	}

	model.ChainId = chainId
	model.Address = address
	model.Hash = hash
	model.Asset = constant.TON
	model.Category = "ton"
	model.Amount = utils.CalculateBalance(amountInt, 9) // nanotons
	model.Status = "Success"
	model.BlockTimestamp = tsMs

	if strings.EqualFold(msg.Source, address) {
		model.Type = "Send"
	} else if strings.EqualFold(msg.Destination, address) {
		model.Type = "Received"
	} else {
		return model, fmt.Errorf("address not involved")
	}
	return model, nil
}
