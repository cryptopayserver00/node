package service

import (
	"context"
	"fmt"
	"math/big"
	"net/url"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
	"node/utils"
	"sort"
	"strings"
	"time"

	"github.com/xssnick/tonutils-go/address"
)

type tonCenterTxResponse struct {
	Ok     bool `json:"ok"`
	Result []struct {
		TransactionID struct {
			Hash string `json:"hash"`
			Lt   string `json:"lt"`
		} `json:"transaction_id"`
		Utime   int64          `json:"utime"`
		Fee     string         `json:"fee"`
		InMsg   *tonCenterMsg  `json:"in_msg"`
		OutMsgs []tonCenterMsg `json:"out_msgs"`
	} `json:"result"`
	Error string `json:"error"`
}

type tonCenterMsg struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Value       string `json:"value"`
	Message     string `json:"message"`
	MsgData     any    `json:"msg_data"`
}

func (n *NService) GetTonTransactions(ctx context.Context, req request.GetTonTransactions) ([]response.ClientTransaction, error) {
	txs, err := n.fetchTonCenterTransactions(ctx, req.ChainId, req.Address)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("ton txs failed chain=%d addr=%s: %v", req.ChainId, req.Address, err))
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

func (n *NService) fetchTonCenterTransactions(ctx context.Context, chainId uint, addressRaw string) ([]response.ClientTransaction, error) {
	base := constant.GetTonCenterAPIBase(chainId)
	if base == "" {
		return nil, fmt.Errorf("unsupported ton chain %d", chainId)
	}

	addr := strings.TrimSpace(addressRaw)
	if addr == "" {
		return nil, fmt.Errorf("empty ton address")
	}

	q := url.Values{}
	q.Set("address", addr)
	q.Set("limit", "50")
	q.Set("archival", "true")
	if key := strings.TrimSpace(global.NODE_CONFIG.Key.ToncenterApiKey); key != "" {
		q.Set("api_key", key)
	}

	httpClient := client
	httpClient.URL = fmt.Sprintf("%s/getTransactions?%s", base, q.Encode())
	httpClient.Timeout = 25 * time.Second
	httpClient.Headers = map[string]string{}
	if key := strings.TrimSpace(global.NODE_CONFIG.Key.ToncenterApiKey); key != "" {
		httpClient.Headers["X-API-Key"] = key
	}

	var resp tonCenterTxResponse
	if err := httpClient.HTTPGet(ctx, &resp); err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("toncenter error: %s", firstNonEmpty(resp.Error, "request failed"))
	}

	seen := make(map[string]struct{}, len(resp.Result)*2)
	var out []response.ClientTransaction
	for _, row := range resp.Result {
		hash := strings.TrimSpace(row.TransactionID.Hash)
		if hash == "" {
			continue
		}
		tsMs := int(row.Utime * 1000)

		if row.InMsg != nil {
			if model, ok := decodeTonMsg(chainId, addr, hash, tsMs, *row.InMsg); ok {
				key := hash + "|" + model.Type + "|" + model.Amount
				if _, exists := seen[key]; !exists {
					seen[key] = struct{}{}
					out = append(out, model)
				}
			}
		}

		for _, msg := range row.OutMsgs {
			if model, ok := decodeTonMsg(chainId, addr, hash, tsMs, msg); ok {
				key := hash + "|" + model.Type + "|" + model.Amount
				if _, exists := seen[key]; !exists {
					seen[key] = struct{}{}
					out = append(out, model)
				}
			}
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].BlockTimestamp > out[j].BlockTimestamp
	})
	return out, nil
}

func decodeTonMsg(chainId uint, walletAddr, hash string, tsMs int, msg tonCenterMsg) (response.ClientTransaction, bool) {
	var model response.ClientTransaction
	amountInt, ok := new(big.Int).SetString(strings.TrimSpace(msg.Value), 10)
	if !ok || amountInt.Sign() <= 0 {
		return model, false
	}

	isSend := tonAddrEqual(msg.Source, walletAddr)
	isRecv := tonAddrEqual(msg.Destination, walletAddr)
	if !isSend && !isRecv {
		return model, false
	}
	// External inbound msgs can have empty source; treat as received when destination matches.
	if isSend && isRecv {
		// Self-transfer / bounce edge: prefer Send when source matches.
		isRecv = false
	}

	model.ChainId = chainId
	model.Address = walletAddr
	model.Hash = hash
	model.Asset = constant.TON
	model.Category = "ton"
	model.Amount = utils.CalculateBalance(amountInt, 9) // nanotons
	model.Status = "Success"
	model.BlockTimestamp = tsMs
	if isSend {
		model.Type = "Send"
	} else {
		model.Type = "Received"
	}
	return model, true
}

// tonAddrEqual compares TON addresses by workchain+account id so 0Q/EQ/UQ forms match.
func tonAddrEqual(a, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == "" || b == "" {
		return false
	}
	aa, errA := address.ParseAddr(a)
	bb, errB := address.ParseAddr(b)
	if errA != nil || errB != nil {
		return strings.EqualFold(a, b)
	}
	return aa.Equals(bb)
}
