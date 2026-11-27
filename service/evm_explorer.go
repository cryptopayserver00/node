package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"node/global"
	"node/global/constant"
	"node/model/node/response"
	sweepUtils "node/sweep/utils"
	"node/utils"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type explorerAPIResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

type explorerNormalTx struct {
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	Hash              string `json:"hash"`
	From              string `json:"from"`
	To                string `json:"to"`
	Value             string `json:"value"`
	IsError           string `json:"isError"`
	TxReceiptStatus   string `json:"txreceipt_status"`
	ContractAddress   string `json:"contractAddress"`
	Input             string `json:"input"`
	FunctionName      string `json:"functionName"`
	GasUsed           string `json:"gasUsed"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
}

type explorerTokenTx struct {
	BlockNumber     string `json:"blockNumber"`
	TimeStamp       string `json:"timeStamp"`
	Hash            string `json:"hash"`
	From            string `json:"from"`
	To              string `json:"to"`
	Value           string `json:"value"`
	TokenName       string `json:"tokenName"`
	TokenSymbol     string `json:"tokenSymbol"`
	TokenDecimal    string `json:"tokenDecimal"`
	ContractAddress string `json:"contractAddress"`
}

// GetEvmTransactionsByExplorer fetches native + ERC20 transfers via
// etherscan-compatible explorers (Etherscan V2 → Blockscout → Routescan).
// Ordinary JSON-RPC cannot list address history; an indexer API is required.
func (n *NService) GetEvmTransactionsByExplorer(
	ctx context.Context,
	chainId uint,
	address string,
	asset string,
) ([]response.ClientTransaction, error) {
	if len(constant.GetExplorerAPIBasesByChainId(chainId)) == 0 {
		return nil, fmt.Errorf("no explorer configured for chain %d", chainId)
	}

	cacheKey := fmt.Sprintf(constantTransactionHistory, chainId, strings.ToLower(address))
	if memcacheEnabled() {
		if item, err := global.NODE_MEMCACHE.Get(cacheKey); err == nil && item != nil {
			var cached []response.ClientTransaction
			if json.Unmarshal(item.Value, &cached) == nil && len(cached) > 0 {
				return filterAndSortEvmTxs(cached, asset), nil
			}
		}
	}

	txs, err := n.fetchEvmExplorerTransactions(ctx, chainId, address)
	if err != nil {
		return nil, err
	}

	if memcacheEnabled() && len(txs) > 0 {
		if raw, mErr := json.Marshal(txs); mErr == nil {
			_ = global.NODE_MEMCACHE.Set(&memcache.Item{
				Key:        cacheKey,
				Value:      raw,
				Expiration: 300, // longer TTL to reduce explorer 429 storms
			})
		}
	}

	return filterAndSortEvmTxs(txs, asset), nil
}

func (n *NService) fetchEvmExplorerTransactions(
	ctx context.Context,
	chainId uint,
	address string,
) ([]response.ClientTransaction, error) {
	bases := constant.GetExplorerAPIBasesByChainId(chainId)
	if len(bases) == 0 {
		return nil, errors.New("no explorer base configured")
	}

	var lastErr error
	for _, base := range bases {
		all, err := n.fetchEvmExplorerTransactionsFromBase(ctx, chainId, address, base)
		if err != nil {
			lastErr = err
			global.NODE_LOG.Error(fmt.Sprintf("explorer chain=%d base=%s: %v", chainId, base, err))
			continue
		}
		return all, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("explorer returned no usable transaction data")
}

func (n *NService) fetchEvmExplorerTransactionsFromBase(
	ctx context.Context,
	chainId uint,
	address string,
	base string,
) ([]response.ClientTransaction, error) {
	var all []response.ClientTransaction

	normalURL := constant.BuildExplorerTxListURL(base, address, 1, 50)
	tokenURL := constant.BuildExplorerTokenTxURL(base, address, 1, 50)

	normalTxs, normalErr := n.getExplorerNormalTxs(ctx, normalURL)
	if normalErr != nil {
		global.NODE_LOG.Error(fmt.Sprintf("explorer txlist chain=%d: %v", chainId, normalErr))
	} else {
		for _, tx := range normalTxs {
			model, dErr := decodeExplorerNormalTx(chainId, address, tx)
			if dErr != nil || model.Hash == "" {
				continue
			}
			all = append(all, model)
		}
	}

	// Brief pause between list calls to stay under free-tier burst limits.
	time.Sleep(200 * time.Millisecond)

	tokenTxs, tokenErr := n.getExplorerTokenTxs(ctx, tokenURL)
	if tokenErr != nil {
		global.NODE_LOG.Error(fmt.Sprintf("explorer tokentx chain=%d: %v", chainId, tokenErr))
	} else {
		for _, tx := range tokenTxs {
			model, dErr := decodeExplorerTokenTx(chainId, address, tx)
			if dErr != nil || model.Hash == "" {
				continue
			}
			all = append(all, model)
		}
	}

	// Dedupe across txlist + tokentx (and identical rows from explorers).
	all = filterAndSortEvmTxs(all, "")

	if normalErr != nil && tokenErr != nil {
		return nil, fmt.Errorf("txlist=%v; tokentx=%v", normalErr, tokenErr)
	}
	// Empty history is valid (new address) when at least one call succeeded.
	if normalErr != nil && tokenErr == nil && len(all) == 0 {
		return all, nil
	}
	if tokenErr != nil && normalErr == nil && len(all) == 0 {
		return all, nil
	}
	return all, nil
}

func (n *NService) getExplorerNormalTxs(ctx context.Context, url string) ([]explorerNormalTx, error) {
	if url == "" {
		return nil, errors.New("empty explorer url")
	}
	httpClient := client
	httpClient.URL = url
	// Keep short so multi-base failover stays responsive under bad networks.
	httpClient.Timeout = 10 * time.Second

	var resp explorerAPIResponse
	if err := httpClient.HTTPGet(ctx, &resp); err != nil {
		return nil, err
	}
	return decodeExplorerResult[explorerNormalTx](resp)
}

func (n *NService) getExplorerTokenTxs(ctx context.Context, url string) ([]explorerTokenTx, error) {
	if url == "" {
		return nil, errors.New("empty explorer url")
	}
	httpClient := client
	httpClient.URL = url
	httpClient.Timeout = 10 * time.Second

	var resp explorerAPIResponse
	if err := httpClient.HTTPGet(ctx, &resp); err != nil {
		return nil, err
	}
	return decodeExplorerResult[explorerTokenTx](resp)
}

func decodeExplorerResult[T any](resp explorerAPIResponse) ([]T, error) {
	msg := strings.TrimSpace(resp.Message)
	status := strings.TrimSpace(resp.Status)

	// Empty / null result: only treat as "no txs" when the explorer says so.
	if len(resp.Result) == 0 || string(resp.Result) == "null" {
		if status == "0" && !isExplorerEmptyHistoryMessage(msg) {
			return []T{}, fmt.Errorf("explorer error: %s", firstNonEmpty(msg, "empty result"))
		}
		return []T{}, nil
	}

	// Explorers return a string in Result when rate-limited / errored / plan-gated.
	if resp.Result[0] == '"' {
		var resultMsg string
		_ = json.Unmarshal(resp.Result, &resultMsg)
		if status == "0" && !isExplorerEmptyHistoryMessage(firstNonEmpty(resultMsg, msg)) {
			return []T{}, fmt.Errorf("explorer error: %s", firstNonEmpty(resultMsg, msg))
		}
		return []T{}, nil
	}

	var rows []T
	if err := json.Unmarshal(resp.Result, &rows); err != nil {
		return nil, err
	}
	// Some explorers return status=0 + [] with "No transactions found".
	if status == "0" && len(rows) == 0 && isExplorerEmptyHistoryMessage(msg) {
		return rows, nil
	}
	if status == "0" && len(rows) == 0 && msg != "" && !strings.EqualFold(msg, "OK") {
		return []T{}, fmt.Errorf("explorer error: %s", msg)
	}
	return rows, nil
}

func isExplorerEmptyHistoryMessage(msg string) bool {
	m := strings.ToLower(strings.TrimSpace(msg))
	if m == "" {
		return false
	}
	if m == "ok" {
		return true
	}
	return strings.Contains(m, "no transaction") ||
		strings.Contains(m, "no records") ||
		strings.Contains(m, "no txs") ||
		strings.EqualFold(m, "no transactions found")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func decodeExplorerNormalTx(chainId uint, address string, tx explorerNormalTx) (response.ClientTransaction, error) {
	var model response.ClientTransaction
	if tx.Hash == "" {
		return model, errors.New("empty hash")
	}

	// Skip pure contract calls with zero value — not useful as payment history.
	if isZeroAmount(tx.Value) {
		return model, errors.New("skip zero-value call")
	}

	model.ChainId = chainId
	model.Address = address
	model.Hash = tx.Hash
	model.Asset = constant.GetNativeAssetByChainId(chainId)
	model.Category = "external"

	amountInt, ok := new(big.Int).SetString(tx.Value, 10)
	if !ok {
		return model, errors.New("invalid value")
	}
	model.Amount = utils.CalculateBalance(amountInt, 18)

	if strings.EqualFold(tx.From, address) {
		model.Type = "Send"
	} else if strings.EqualFold(tx.To, address) {
		model.Type = "Received"
	} else {
		return model, errors.New("address not involved")
	}

	if tx.IsError == "1" || tx.TxReceiptStatus == "0" {
		model.Status = "Failed"
	} else {
		model.Status = "Success"
	}
	model.BlockTimestamp = explorerTimestampMs(tx.TimeStamp)
	return model, nil
}

func decodeExplorerTokenTx(chainId uint, address string, tx explorerTokenTx) (response.ClientTransaction, error) {
	var model response.ClientTransaction
	if tx.Hash == "" {
		return model, errors.New("empty hash")
	}

	model.ChainId = chainId
	model.Address = address
	model.Hash = tx.Hash
	model.Category = "erc20"
	model.ContractAddress = tx.ContractAddress

	decimals := 18
	if d, err := strconv.Atoi(tx.TokenDecimal); err == nil && d >= 0 && d <= 36 {
		decimals = d
	}

	asset := strings.TrimSpace(tx.TokenSymbol)
	if ok, name, contract, regDecimals := sweepUtils.GetContractInfo(chainId, tx.ContractAddress); ok {
		asset = name
		model.ContractAddress = contract
		decimals = regDecimals
	}
	if asset == "" {
		asset = "TOKEN"
	}
	model.Asset = asset

	amountInt, ok := new(big.Int).SetString(tx.Value, 10)
	if !ok {
		return model, errors.New("invalid token value")
	}
	model.Amount = utils.CalculateBalance(amountInt, decimals)

	if strings.EqualFold(tx.From, address) {
		model.Type = "Send"
	} else if strings.EqualFold(tx.To, address) {
		model.Type = "Received"
	} else {
		return model, errors.New("address not involved")
	}

	model.Status = "Success"
	model.BlockTimestamp = explorerTimestampMs(tx.TimeStamp)
	return model, nil
}

func explorerTimestampMs(ts string) int {
	sec, err := strconv.ParseInt(ts, 10, 64)
	if err != nil || sec <= 0 {
		return 0
	}
	return int(sec * 1000)
}

func isZeroAmount(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" || v == "0" {
		return true
	}
	n, ok := new(big.Int).SetString(v, 10)
	return ok && n.Sign() == 0
}

func normalizeTxAmount(amount string) string {
	amount = strings.TrimSpace(amount)
	if amount == "" {
		return "0"
	}
	// Strip trailing zeros so "0.001000" and "0.001" collide.
	if strings.Contains(amount, ".") {
		amount = strings.TrimRight(amount, "0")
		amount = strings.TrimRight(amount, ".")
	}
	if amount == "" || amount == "-" {
		return "0"
	}
	return amount
}

func normalizeTxCategory(category string) string {
	c := strings.ToLower(strings.TrimSpace(category))
	switch c {
	case "eth", "external", "native", "":
		return "native"
	case "erc20", "token":
		return "erc20"
	default:
		return c
	}
}

func clientTxDedupeKey(tx response.ClientTransaction) string {
	// Identity for wallet history: hash + asset + direction + amount.
	// Do NOT key on contractAddress — explorers/Alchemy often fill it inconsistently
	// for the same transfer (empty vs token contract), which created false duplicates.
	// Distinct transfers in one tx usually differ by asset and/or amount (e.g. swaps).
	return strings.ToLower(strings.TrimSpace(tx.Hash)) + "|" +
		strings.ToLower(strings.TrimSpace(tx.Asset)) + "|" +
		strings.ToLower(strings.TrimSpace(tx.Type)) + "|" +
		normalizeTxAmount(tx.Amount) + "|" +
		normalizeTxCategory(tx.Category)
}

func preferClientTx(a, b response.ClientTransaction) response.ClientTransaction {
	// Prefer confirmed over pending/failed when the same transfer appears twice.
	rank := func(status string) int {
		if strings.EqualFold(status, "Success") {
			return 3
		}
		if strings.EqualFold(status, "Pending") {
			return 1
		}
		if strings.EqualFold(status, "Failed") {
			return 0
		}
		return 2
	}
	chosen := a
	if rank(b.Status) > rank(a.Status) {
		chosen = b
	} else if rank(b.Status) == rank(a.Status) && b.BlockTimestamp > a.BlockTimestamp {
		chosen = b
	}
	// Keep whichever contractAddress is non-empty for ERC20 display.
	if strings.TrimSpace(chosen.ContractAddress) == "" {
		if strings.TrimSpace(a.ContractAddress) != "" {
			chosen.ContractAddress = a.ContractAddress
		} else if strings.TrimSpace(b.ContractAddress) != "" {
			chosen.ContractAddress = b.ContractAddress
		}
	}
	return chosen
}

func filterAndSortEvmTxs(txs []response.ClientTransaction, asset string) []response.ClientTransaction {
	asset = strings.TrimSpace(asset)
	seen := make(map[string]response.ClientTransaction, len(txs))
	order := make([]string, 0, len(txs))

	for _, tx := range txs {
		if tx.Hash == "" {
			continue
		}
		if asset != "" && !strings.EqualFold(tx.Asset, asset) {
			continue
		}
		key := clientTxDedupeKey(tx)
		if existing, ok := seen[key]; ok {
			seen[key] = preferClientTx(existing, tx)
			continue
		}
		seen[key] = tx
		order = append(order, key)
	}

	out := make([]response.ClientTransaction, 0, len(order))
	for _, key := range order {
		out = append(out, seen[key])
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].BlockTimestamp == 0 && out[j].BlockTimestamp != 0 {
			return true
		}
		if out[i].BlockTimestamp != 0 && out[j].BlockTimestamp == 0 {
			return false
		}
		return out[i].BlockTimestamp > out[j].BlockTimestamp
	})
	return out
}
