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

// GetEvmTransactionsByExplorer fetches native + ERC20 transfers via free
// etherscan-compatible explorers (Blockscout / Routescan). Results are returned
// directly and optionally cached in memcache when enabled.
func (n *NService) GetEvmTransactionsByExplorer(
	ctx context.Context,
	chainId uint,
	address string,
	asset string,
) ([]response.ClientTransaction, error) {
	if constant.GetExplorerAPIBaseByChainId(chainId) == "" {
		return nil, fmt.Errorf("no free explorer configured for chain %d", chainId)
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
				Expiration: 120, // short TTL — explorer data changes often
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
	var all []response.ClientTransaction

	normalURL := constant.GetExplorerTxListURL(chainId, address, 1, 50)
	tokenURL := constant.GetExplorerTokenTxURL(chainId, address, 1, 50)

	normalTxs, err := n.getExplorerNormalTxs(ctx, normalURL)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("explorer txlist chain=%d: %v", chainId, err))
	} else {
		for _, tx := range normalTxs {
			model, dErr := decodeExplorerNormalTx(chainId, address, tx)
			if dErr != nil || model.Hash == "" {
				continue
			}
			all = append(all, model)
		}
	}

	tokenTxs, err := n.getExplorerTokenTxs(ctx, tokenURL)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("explorer tokentx chain=%d: %v", chainId, err))
	} else {
		for _, tx := range tokenTxs {
			model, dErr := decodeExplorerTokenTx(chainId, address, tx)
			if dErr != nil || model.Hash == "" {
				continue
			}
			all = append(all, model)
		}
	}

	if len(all) == 0 && normalTxs == nil && tokenTxs == nil {
		return nil, errors.New("explorer returned no usable transaction data")
	}

	return all, nil
}

func (n *NService) getExplorerNormalTxs(ctx context.Context, url string) ([]explorerNormalTx, error) {
	if url == "" {
		return nil, errors.New("empty explorer url")
	}
	httpClient := client
	httpClient.URL = url
	httpClient.Timeout = 20 * time.Second

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
	httpClient.Timeout = 20 * time.Second

	var resp explorerAPIResponse
	if err := httpClient.HTTPGet(ctx, &resp); err != nil {
		return nil, err
	}
	return decodeExplorerResult[explorerTokenTx](resp)
}

func decodeExplorerResult[T any](resp explorerAPIResponse) ([]T, error) {
	if len(resp.Result) == 0 || string(resp.Result) == "null" {
		return []T{}, nil
	}
	// Explorers return a string in Result when rate-limited / errored.
	if resp.Result[0] == '"' {
		var msg string
		_ = json.Unmarshal(resp.Result, &msg)
		if resp.Status == "0" {
			return []T{}, fmt.Errorf("explorer error: %s", firstNonEmpty(msg, resp.Message))
		}
		return []T{}, nil
	}
	var rows []T
	if err := json.Unmarshal(resp.Result, &rows); err != nil {
		return nil, err
	}
	return rows, nil
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

func filterAndSortEvmTxs(txs []response.ClientTransaction, asset string) []response.ClientTransaction {
	asset = strings.TrimSpace(asset)
	out := make([]response.ClientTransaction, 0, len(txs))
	for _, tx := range txs {
		if asset == "" || strings.EqualFold(tx.Asset, asset) {
			out = append(out, tx)
		}
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
