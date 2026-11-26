package service

import (
	"context"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
	"sort"
	"time"
)

type solanaSignature struct {
	Signature string `json:"signature"`
	Slot      uint64 `json:"slot"`
	BlockTime *int64 `json:"blockTime"`
	Err       any    `json:"err"`
}

type solanaSignaturesRPCResponse struct {
	Result []solanaSignature `json:"result"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// GetSolanaTransactions returns recent signatures for an address via public RPC.
// Full parsed instruction decoding remains on the cryptopayserver client; here we
// expose a lightweight history usable by wallet UIs and monitoring.
func (n *NService) GetSolanaTransactions(ctx context.Context, req request.GetSolanaTransactions) ([]response.ClientTransaction, error) {
	sigs, err := n.fetchSolanaSignatures(ctx, req.ChainId, req.Address, 40)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("solana txs failed: %v", err))
		return []response.ClientTransaction{}, nil
	}

	out := make([]response.ClientTransaction, 0, len(sigs))
	for _, s := range sigs {
		model := response.ClientTransaction{
			ChainId: req.ChainId,
			Address: req.Address,
			Hash:    s.Signature,
			Asset:   constant.SOL,
			Type:    "Transfer",
			Category: "sol",
			Status:  "Success",
			Amount:  "",
		}
		if s.Err != nil {
			model.Status = "Failed"
		}
		if s.BlockTime != nil {
			model.BlockTimestamp = int((*s.BlockTime) * 1000)
		}
		out = append(out, model)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].BlockTimestamp > out[j].BlockTimestamp
	})
	return out, nil
}

func (n *NService) GetSolTransactions(ctx context.Context, req request.GetSolTransactions) ([]response.ClientTransaction, error) {
	return n.GetSolanaTransactions(ctx, request.GetSolanaTransactions{
		ChainId: req.ChainId,
		Address: req.Address,
	})
}

func (n *NService) GetSplTransactions(ctx context.Context, req request.GetSplTransactions) ([]response.ClientTransaction, error) {
	// SPL token parsing needs getParsedTransactions; signatures for the token account
	// can still be listed when contract address is actually a token account.
	_ = req.ContractAddress
	return n.GetSolanaTransactions(ctx, request.GetSolanaTransactions{
		ChainId: req.ChainId,
		Address: req.Address,
	})
}

func (n *NService) fetchSolanaSignatures(ctx context.Context, chainId uint, address string, limit int) ([]solanaSignature, error) {
	rpc := constant.GetRPCUrlByChainId(chainId)
	if rpc == "" {
		return nil, fmt.Errorf("no solana rpc for chain %d", chainId)
	}
	if limit <= 0 {
		limit = 20
	}

	httpClient := client
	httpClient.URL = rpc
	httpClient.Timeout = 20 * time.Second

	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getSignaturesForAddress",
		"params": []any{
			address,
			map[string]any{"limit": limit},
		},
	}

	var resp solanaSignaturesRPCResponse
	if err := httpClient.HTTPPost(ctx, payload, &resp); err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("solana rpc: %s", resp.Error.Message)
	}
	return resp.Result, nil
}
