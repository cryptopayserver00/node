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

// trezor/blockbook style address txs (free public endpoints)
type bchAddressTxResponse struct {
	Transactions []struct {
		TxID          string `json:"txid"`
		BlockTime     int64  `json:"blockTime"`
		Confirmations int64  `json:"confirmations"`
		Vin           []struct {
			Addresses []string `json:"addresses"`
			Value     string   `json:"value"`
		} `json:"vin"`
		Vout []struct {
			Addresses []string `json:"addresses"`
			Value     string   `json:"value"`
		} `json:"vout"`
	} `json:"transactions"`
}

func (n *NService) GetBchTransactions(ctx context.Context, req request.GetBchTransactions) ([]response.ClientTransaction, error) {
	txs, err := n.fetchBchTransactions(ctx, req.ChainId, req.Address)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("bch txs failed: %v", err))
		return []response.ClientTransaction{}, nil
	}
	return filterAndSortEvmTxs(txs, req.Asset), nil
}

func bchExplorerBase(chainId uint) string {
	switch chainId {
	case constant.BCH_MAINNET:
		return "https://bch1.trezor.io/api/v2"
	case constant.BCH_TESTNET:
		// Limited public coverage; try blockchair-compatible fallback path if available.
		return "https://bchtest1.trezor.io/api/v2"
	default:
		return ""
	}
}

func (n *NService) fetchBchTransactions(ctx context.Context, chainId uint, address string) ([]response.ClientTransaction, error) {
	base := bchExplorerBase(chainId)
	if base == "" {
		return nil, fmt.Errorf("unsupported bch chain %d", chainId)
	}

	httpClient := client
	httpClient.URL = fmt.Sprintf("%s/address/%s?details=txs&pageSize=50", base, address)
	httpClient.Timeout = 20 * time.Second

	var resp bchAddressTxResponse
	if err := httpClient.HTTPGet(ctx, &resp); err != nil {
		return nil, err
	}

	var out []response.ClientTransaction
	for _, tx := range resp.Transactions {
		sent := false
		received := false
		var sentVal, recvVal int64

		for _, vin := range tx.Vin {
			for _, a := range vin.Addresses {
				if strings.EqualFold(a, address) || a == address {
					sent = true
					sentVal += parseLooseInt64(vin.Value)
				}
			}
		}
		for _, vout := range tx.Vout {
			for _, a := range vout.Addresses {
				if strings.EqualFold(a, address) || a == address {
					received = true
					recvVal += parseLooseInt64(vout.Value)
				}
			}
		}

		model := response.ClientTransaction{
			ChainId:        chainId,
			Address:        address,
			Hash:           tx.TxID,
			Asset:          constant.BCH,
			Category:       "bch",
			Status:         "Success",
			BlockTimestamp: int(tx.BlockTime * 1000),
		}
		if tx.Confirmations == 0 {
			model.Status = "Pending"
		}

		switch {
		case sent && !received:
			model.Type = "Send"
			model.Amount = formatSatoshis(sentVal)
		case received && !sent:
			model.Type = "Received"
			model.Amount = formatSatoshis(recvVal)
		case sent && received:
			// Change-back transaction: net send approx
			net := sentVal - recvVal
			if net < 0 {
				net = -net
				model.Type = "Received"
			} else {
				model.Type = "Send"
			}
			model.Amount = formatSatoshis(net)
		default:
			continue
		}
		if model.Amount == "" || model.Amount == "0" {
			continue
		}
		out = append(out, model)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].BlockTimestamp > out[j].BlockTimestamp
	})
	return out, nil
}

func parseLooseInt64(v string) int64 {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	var n int64
	for _, c := range v {
		if c < '0' || c > '9' {
			continue
		}
		n = n*10 + int64(c-'0')
	}
	return n
}

func formatSatoshis(sats int64) string {
	if sats <= 0 {
		return "0"
	}
	whole := sats / 1e8
	frac := sats % 1e8
	if frac == 0 {
		return fmt.Sprintf("%d", whole)
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%d.%08d", whole, frac), "0"), ".")
}
