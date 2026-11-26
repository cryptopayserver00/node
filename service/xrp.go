package service

import (
	"context"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
	"sort"
	"strings"
	"time"
)

type xrpAccountTxResponse struct {
	Result struct {
		Status       string `json:"status"`
		Account      string `json:"account"`
		Transactions []struct {
			Meta struct {
				TransactionResult string `json:"TransactionResult"`
				DeliveredAmount   any    `json:"delivered_amount"`
			} `json:"meta"`
			Tx struct {
				Hash            string `json:"hash"`
				Account         string `json:"Account"`
				Destination     string `json:"Destination"`
				TransactionType string `json:"TransactionType"`
				Amount          any    `json:"Amount"`
				Date            int64  `json:"date"`
			} `json:"tx"`
			Validated bool `json:"validated"`
		} `json:"transactions"`
	} `json:"result"`
}

func (n *NService) GetXrpTransactions(ctx context.Context, req request.GetXrpTransactions) ([]response.ClientTransaction, error) {
	txs, err := n.fetchXrpAccountTx(ctx, req.ChainId, req.Address)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("xrp txs failed: %v", err))
		return []response.ClientTransaction{}, nil
	}
	return filterAndSortEvmTxs(txs, req.Asset), nil
}

func (n *NService) fetchXrpAccountTx(ctx context.Context, chainId uint, address string) ([]response.ClientTransaction, error) {
	rpc := constant.GetRPCUrlByChainId(chainId)
	if rpc == "" {
		return nil, fmt.Errorf("no xrp rpc for chain %d", chainId)
	}

	httpClient := client
	httpClient.URL = rpc
	httpClient.Timeout = 20 * time.Second

	payload := map[string]any{
		"method": "account_tx",
		"params": []map[string]any{
			{
				"account":        address,
				"ledger_index_min": -1,
				"ledger_index_max": -1,
				"limit":          50,
				"forward":        false,
			},
		},
	}

	var resp xrpAccountTxResponse
	if err := httpClient.HTTPPost(ctx, payload, &resp); err != nil {
		return nil, err
	}
	if !strings.EqualFold(resp.Result.Status, "success") {
		return nil, fmt.Errorf("xrp rpc status=%s", resp.Result.Status)
	}

	var out []response.ClientTransaction
	for _, row := range resp.Result.Transactions {
		if row.Tx.TransactionType != "Payment" {
			continue
		}
		amount := xrpAmountToXRP(row.Meta.DeliveredAmount)
		if amount == "" {
			amount = xrpAmountToXRP(row.Tx.Amount)
		}
		if amount == "" || amount == "0" {
			continue
		}

		model := response.ClientTransaction{
			ChainId:        chainId,
			Address:        address,
			Hash:           row.Tx.Hash,
			Amount:         amount,
			Asset:          constant.XRP,
			Category:       "xrp",
			Status:         "Success",
			BlockTimestamp: xrpRippleTimeToMs(row.Tx.Date),
		}
		if row.Meta.TransactionResult != "" && row.Meta.TransactionResult != "tesSUCCESS" {
			model.Status = "Failed"
		}
		if strings.EqualFold(row.Tx.Account, address) {
			model.Type = "Send"
		} else if strings.EqualFold(row.Tx.Destination, address) {
			model.Type = "Received"
		} else {
			continue
		}
		out = append(out, model)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].BlockTimestamp > out[j].BlockTimestamp
	})
	return out, nil
}

func xrpAmountToXRP(v any) string {
	switch t := v.(type) {
	case string:
		// drops
		return utilsDropsToXRP(t)
	case map[string]any:
		// IOU — skip non-XRP for now
		return ""
	default:
		return ""
	}
}

func utilsDropsToXRP(drops string) string {
	drops = strings.TrimSpace(drops)
	if drops == "" {
		return ""
	}
	// 1 XRP = 1_000_000 drops
	var whole, frac string
	if len(drops) <= 6 {
		whole = "0"
		frac = strings.Repeat("0", 6-len(drops)) + drops
	} else {
		whole = drops[:len(drops)-6]
		frac = drops[len(drops)-6:]
	}
	frac = strings.TrimRight(frac, "0")
	if frac == "" {
		return whole
	}
	return whole + "." + frac
}

// XRPL epoch starts 2000-01-01; unix = ripple + 946684800
func xrpRippleTimeToMs(rippleDate int64) int {
	if rippleDate <= 0 {
		return 0
	}
	return int((rippleDate + 946684800) * 1000)
}
