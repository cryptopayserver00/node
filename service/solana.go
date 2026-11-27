package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
	sweepUtils "node/sweep/utils"
	"node/utils"
	"sort"
	"strings"
	"time"
)

type solanaSignature struct {
	Signature string `json:"signature"`
	Slot      uint64 `json:"slot"`
	BlockTime *int64 `json:"blockTime"`
	Err       any    `json:"err"`
}

type solanaRPCError struct {
	Message string `json:"message"`
}

type solanaSignaturesRPCResponse struct {
	Result []solanaSignature `json:"result"`
	Error  *solanaRPCError   `json:"error"`
}

type solanaTxRPCResponse struct {
	Result *solanaParsedTx `json:"result"`
	Error  *solanaRPCError `json:"error"`
}

type solanaParsedTx struct {
	BlockTime *int64 `json:"blockTime"`
	Meta      *struct {
		Err               any                 `json:"err"`
		Fee               uint64              `json:"fee"`
		PreBalances       []uint64            `json:"preBalances"`
		PostBalances      []uint64            `json:"postBalances"`
		PreTokenBalances  []solanaTokenBal    `json:"preTokenBalances"`
		PostTokenBalances []solanaTokenBal    `json:"postTokenBalances"`
	} `json:"meta"`
	Transaction struct {
		Signatures []string `json:"signatures"`
		Message    struct {
			AccountKeys  []json.RawMessage `json:"accountKeys"`
			Instructions []json.RawMessage `json:"instructions"`
		} `json:"message"`
	} `json:"transaction"`
}

type solanaTokenBal struct {
	AccountIndex  int    `json:"accountIndex"`
	Mint          string `json:"mint"`
	Owner         string `json:"owner"`
	UITokenAmount struct {
		Amount         string  `json:"amount"`
		Decimals       int     `json:"decimals"`
		UIAmountString string  `json:"uiAmountString"`
		UIAmount       *float64 `json:"uiAmount"`
	} `json:"uiTokenAmount"`
}

// GetSolanaTransactions returns recent SOL + SPL transfers for an address.
func (n *NService) GetSolanaTransactions(ctx context.Context, req request.GetSolanaTransactions) ([]response.ClientTransaction, error) {
	txs, err := n.fetchSolanaAddressTransactions(ctx, req.ChainId, req.Address, 30)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("solana txs failed chain=%d addr=%s: %v", req.ChainId, req.Address, err))
		return []response.ClientTransaction{}, nil
	}
	return txs, nil
}

func (n *NService) GetSolTransactions(ctx context.Context, req request.GetSolTransactions) ([]response.ClientTransaction, error) {
	txs, err := n.GetSolanaTransactions(ctx, request.GetSolanaTransactions{
		ChainId: req.ChainId,
		Address: req.Address,
	})
	if err != nil {
		return nil, err
	}
	out := make([]response.ClientTransaction, 0, len(txs))
	for _, tx := range txs {
		if strings.EqualFold(tx.Asset, constant.SOL) && tx.ContractAddress == "" {
			out = append(out, tx)
		}
	}
	return out, nil
}

func (n *NService) GetSplTransactions(ctx context.Context, req request.GetSplTransactions) ([]response.ClientTransaction, error) {
	txs, err := n.GetSolanaTransactions(ctx, request.GetSolanaTransactions{
		ChainId: req.ChainId,
		Address: req.Address,
	})
	if err != nil {
		return nil, err
	}
	mint := strings.TrimSpace(req.ContractAddress)
	out := make([]response.ClientTransaction, 0, len(txs))
	for _, tx := range txs {
		if mint != "" && !strings.EqualFold(tx.ContractAddress, mint) {
			continue
		}
		if tx.ContractAddress == "" {
			continue
		}
		out = append(out, tx)
	}
	return out, nil
}

func (n *NService) fetchSolanaAddressTransactions(ctx context.Context, chainId uint, address string, limit int) ([]response.ClientTransaction, error) {
	addr := strings.TrimSpace(address)
	if addr == "" {
		return nil, fmt.Errorf("empty solana address")
	}
	sigs, err := n.fetchSolanaSignatures(ctx, chainId, addr, limit)
	if err != nil {
		return nil, err
	}
	if len(sigs) == 0 {
		return []response.ClientTransaction{}, nil
	}

	sigList := make([]string, 0, len(sigs))
	for _, s := range sigs {
		if strings.TrimSpace(s.Signature) != "" {
			sigList = append(sigList, s.Signature)
		}
	}

	parsed, err := n.fetchSolanaParsedTransactions(ctx, chainId, sigList)
	if err != nil {
		return nil, err
	}

	out := make([]response.ClientTransaction, 0, len(parsed))
	for i, tx := range parsed {
		if tx == nil {
			continue
		}
		sig := ""
		if i < len(sigList) {
			sig = sigList[i]
		}
		if len(tx.Transaction.Signatures) > 0 {
			sig = tx.Transaction.Signatures[0]
		}
		out = append(out, decodeSolanaTxForAddress(chainId, addr, sig, tx)...)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].BlockTimestamp > out[j].BlockTimestamp
	})
	return out, nil
}

func (n *NService) fetchSolanaSignatures(ctx context.Context, chainId uint, address string, limit int) ([]solanaSignature, error) {
	if limit <= 0 {
		limit = 20
	}
	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getSignaturesForAddress",
		"params": []any{
			address,
			map[string]any{"limit": limit},
		},
	}

	var lastErr error
	for _, rpc := range solanaRPCCandidates(chainId) {
		var resp solanaSignaturesRPCResponse
		if err := solanaRPCPost(ctx, chainId, rpc, payload, &resp); err != nil {
			lastErr = err
			continue
		}
		if resp.Error != nil {
			lastErr = fmt.Errorf("solana rpc: %s", resp.Error.Message)
			constant.ReportRPCFailure(chainId, rpc)
			continue
		}
		constant.ReportRPCSuccess(chainId, rpc)
		return resp.Result, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no solana rpc for chain %d", chainId)
	}
	return nil, lastErr
}

func (n *NService) fetchSolanaParsedTransactions(ctx context.Context, chainId uint, signatures []string) ([]*solanaParsedTx, error) {
	if len(signatures) == 0 {
		return nil, nil
	}

	batch := make([]map[string]any, 0, len(signatures))
	for i, sig := range signatures {
		batch = append(batch, map[string]any{
			"jsonrpc": "2.0",
			"id":      i + 1,
			"method":  "getTransaction",
			"params": []any{
				sig,
				map[string]any{
					"encoding":                       "jsonParsed",
					"commitment":                     "confirmed",
					"maxSupportedTransactionVersion": 0,
				},
			},
		})
	}

	var lastErr error
	for _, rpc := range solanaRPCCandidates(chainId) {
		raw, err := solanaRPCPostRaw(ctx, chainId, rpc, batch)
		if err != nil {
			lastErr = err
			continue
		}

		var items []solanaTxRPCResponse
		if err := json.Unmarshal(raw, &items); err != nil {
			// Some RPCs reject batches; fall back to sequential.
			items = make([]solanaTxRPCResponse, len(signatures))
			okAny := false
			for i, sig := range signatures {
				payload := map[string]any{
					"jsonrpc": "2.0",
					"id":      1,
					"method":  "getTransaction",
					"params": []any{
						sig,
						map[string]any{
							"encoding":                       "jsonParsed",
							"commitment":                     "confirmed",
							"maxSupportedTransactionVersion": 0,
						},
					},
				}
				var one solanaTxRPCResponse
				if err := solanaRPCPost(ctx, chainId, rpc, payload, &one); err != nil {
					lastErr = err
					continue
				}
				items[i] = one
				okAny = true
			}
			if !okAny {
				continue
			}
		}

		out := make([]*solanaParsedTx, len(signatures))
		for _, item := range items {
			if item.Error != nil {
				continue
			}
			if item.Result == nil {
				continue
			}
			// Match by signature when possible; otherwise fill by response order via id is hard —
			// batch responses may be unordered. Prefer signature match.
			sig := ""
			if len(item.Result.Transaction.Signatures) > 0 {
				sig = item.Result.Transaction.Signatures[0]
			}
			idx := indexOfString(signatures, sig)
			if idx < 0 {
				continue
			}
			out[idx] = item.Result
		}
		constant.ReportRPCSuccess(chainId, rpc)
		return out, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no solana rpc for chain %d", chainId)
	}
	return nil, lastErr
}

func decodeSolanaTxForAddress(chainId uint, address, hash string, tx *solanaParsedTx) []response.ClientTransaction {
	if tx == nil || tx.Meta == nil {
		return nil
	}

	tsMs := 0
	if tx.BlockTime != nil {
		tsMs = int((*tx.BlockTime) * 1000)
	}
	status := "Success"
	if tx.Meta.Err != nil {
		status = "Failed"
	}

	keys := solanaAccountKeys(tx.Transaction.Message.AccountKeys)
	var out []response.ClientTransaction

	// SPL / token balance changes for this owner
	tokenDeltas := solanaTokenDeltasForOwner(address, tx.Meta.PreTokenBalances, tx.Meta.PostTokenBalances)
	for mint, delta := range tokenDeltas {
		if delta.Sign() == 0 {
			continue
		}
		decimals := 0
		uiAmount := ""
		asset := "TOKEN"
		if ok, name, _, regDecimals := sweepUtils.GetContractInfo(chainId, mint); ok {
			asset = name
			decimals = regDecimals
		} else {
			// Fall back to decimals from post/pre token meta
			decimals = solanaTokenDecimals(mint, tx.Meta.PreTokenBalances, tx.Meta.PostTokenBalances)
		}
		abs := new(big.Int).Abs(delta)
		if uiAmount == "" {
			uiAmount = utils.CalculateBalance(abs, decimals)
		}
		model := response.ClientTransaction{
			ChainId:         chainId,
			Address:         address,
			Hash:            hash,
			Amount:          uiAmount,
			Asset:           asset,
			ContractAddress: mint,
			Category:        "spl",
			Status:          status,
			BlockTimestamp:  tsMs,
		}
		if delta.Sign() > 0 {
			model.Type = "Received"
		} else {
			model.Type = "Send"
		}
		out = append(out, model)
	}

	// Native SOL balance change for this address
	idx := indexOfString(keys, address)
	if idx >= 0 && idx < len(tx.Meta.PreBalances) && idx < len(tx.Meta.PostBalances) {
		pre := int64(tx.Meta.PreBalances[idx])
		post := int64(tx.Meta.PostBalances[idx])
		delta := post - pre
		fee := int64(tx.Meta.Fee)
		skipFeeOnly := delta < 0 && -delta == fee && len(tokenDeltas) > 0
		if delta != 0 && !skipFeeOnly {
			abs := delta
			if abs < 0 {
				abs = -abs
			}
			model := response.ClientTransaction{
				ChainId:        chainId,
				Address:        address,
				Hash:           hash,
				Amount:         utils.CalculateBalance(big.NewInt(abs), 9),
				Asset:          constant.SOL,
				Category:       "sol",
				Status:         status,
				BlockTimestamp: tsMs,
			}
			if delta > 0 {
				model.Type = "Received"
			} else {
				model.Type = "Send"
			}
			out = append(out, model)
		}
	}

	return out
}

func solanaTokenDeltasForOwner(owner string, pre, post []solanaTokenBal) map[string]*big.Int {
	preMap := map[string]*big.Int{}
	postMap := map[string]*big.Int{}

	accumulate := func(dst map[string]*big.Int, rows []solanaTokenBal) {
		for _, row := range rows {
			if !strings.EqualFold(strings.TrimSpace(row.Owner), owner) {
				continue
			}
			mint := strings.TrimSpace(row.Mint)
			if mint == "" {
				continue
			}
			v, ok := new(big.Int).SetString(strings.TrimSpace(row.UITokenAmount.Amount), 10)
			if !ok {
				v = big.NewInt(0)
			}
			if cur, exists := dst[mint]; exists {
				dst[mint] = new(big.Int).Add(cur, v)
			} else {
				dst[mint] = v
			}
		}
	}
	accumulate(preMap, pre)
	accumulate(postMap, post)

	mints := map[string]struct{}{}
	for m := range preMap {
		mints[m] = struct{}{}
	}
	for m := range postMap {
		mints[m] = struct{}{}
	}

	out := make(map[string]*big.Int, len(mints))
	for mint := range mints {
		a := preMap[mint]
		b := postMap[mint]
		if a == nil {
			a = big.NewInt(0)
		}
		if b == nil {
			b = big.NewInt(0)
		}
		out[mint] = new(big.Int).Sub(b, a)
	}
	return out
}

func solanaTokenDecimals(mint string, pre, post []solanaTokenBal) int {
	for _, row := range append(append([]solanaTokenBal{}, post...), pre...) {
		if strings.EqualFold(row.Mint, mint) {
			return row.UITokenAmount.Decimals
		}
	}
	return 0
}

func solanaAccountKeys(raw []json.RawMessage) []string {
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		var asString string
		if err := json.Unmarshal(item, &asString); err == nil && asString != "" {
			out = append(out, asString)
			continue
		}
		var asObj struct {
			Pubkey string `json:"pubkey"`
		}
		if err := json.Unmarshal(item, &asObj); err == nil && asObj.Pubkey != "" {
			out = append(out, asObj.Pubkey)
		}
	}
	return out
}

func solanaRPCCandidates(chainId uint) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(u string) {
		u = strings.TrimSpace(u)
		if u == "" {
			return
		}
		if _, ok := seen[u]; ok {
			return
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	add(constant.GetRPCUrlByChainId(chainId))
	for _, u := range constant.GetAllRPCUrlByChainId(chainId) {
		add(u)
	}
	return out
}

func solanaRPCPost(ctx context.Context, chainId uint, rpc string, payload any, dest any) error {
	httpClient := client
	httpClient.URL = rpc
	httpClient.Timeout = 25 * time.Second
	if err := httpClient.HTTPPost(ctx, payload, dest); err != nil {
		constant.ReportRPCFailure(chainId, rpc)
		return err
	}
	return nil
}

func solanaRPCPostRaw(ctx context.Context, chainId uint, rpc string, payload any) ([]byte, error) {
	httpClient := client
	httpClient.URL = rpc
	httpClient.Timeout = 35 * time.Second

	// Reuse HTTPPost path but capture raw via a json.RawMessage wrapper when possible.
	var raw json.RawMessage
	if err := httpClient.HTTPPost(ctx, payload, &raw); err != nil {
		constant.ReportRPCFailure(chainId, rpc)
		return nil, err
	}
	return []byte(raw), nil
}

func indexOfString(list []string, target string) int {
	for i, v := range list {
		if v == target {
			return i
		}
	}
	return -1
}
