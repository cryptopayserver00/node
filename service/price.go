package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/url"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
	"strings"
	"time"
)

// GetCryptoPrice serves cached CoinGecko prices.
// - format=symbol → deshop CustomCoingeckoPrice
// - otherwise → CoinGecko id-keyed map; optional ids/currency filter (CPS compatible)
func (n *NService) GetCryptoPrice(ctx context.Context, req request.GetCryptoPrice) (any, error) {
	if global.NODE_REDIS == nil {
		return nil, errors.New("redis not available")
	}

	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format == "" {
		format = "gecko"
	}
	currency := strings.ToLower(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "usd"
	}

	if format == "symbol" {
		data, err := global.NODE_REDIS.Get(ctx, constant.CRYPTO_PRICE_SYMBOL).Result()
		if err != nil {
			return nil, fmt.Errorf("crypto price cache miss: %w", err)
		}
		var price response.CustomCoingeckoPrice
		if err := json.Unmarshal([]byte(data), &price); err != nil {
			return nil, err
		}
		// deshop expects the symbol-keyed object directly as response data
		return price, nil
	}

	// Non-USD: live-fetch (short-cache) so CPS invoice currencies still work.
	if currency != "usd" {
		return n.fetchLiveGeckoPrices(ctx, req.Ids, currency)
	}

	raw, err := global.NODE_REDIS.Get(ctx, constant.CRYPTO_PRICE_RAW).Result()
	if err != nil {
		return nil, fmt.Errorf("crypto price cache miss: %w", err)
	}

	var generic map[string]any
	if err := json.Unmarshal([]byte(raw), &generic); err != nil {
		return nil, err
	}
	meta := readPriceMeta(ctx)
	return filterGeckoAny(generic, req.Ids, currency, meta.UpdatedAt), nil
}

func readPriceMeta(ctx context.Context) response.CryptoPriceMeta {
	var meta response.CryptoPriceMeta
	data, err := global.NODE_REDIS.Get(ctx, constant.CRYPTO_PRICE_META).Result()
	if err != nil {
		return meta
	}
	_ = json.Unmarshal([]byte(data), &meta)
	return meta
}

func filterGeckoAny(all map[string]any, ids, currency string, updatedAt int64) map[string]any {
	idList := splitCSV(ids)
	out := make(map[string]any)
	selectID := func(id string, v any) {
		m, ok := v.(map[string]any)
		if !ok {
			out[id] = v
			return
		}
		// Keep currency fields + last_updated_at (CoinGecko-compatible)
		picked := make(map[string]any)
		for k, val := range m {
			lk := strings.ToLower(k)
			if lk == "last_updated_at" || lk == currency || strings.HasPrefix(lk, currency+"_") {
				picked[lk] = val
			}
		}
		if len(picked) == 0 {
			out[id] = m
		} else {
			out[id] = picked
		}
	}

	if len(idList) == 0 {
		for id, v := range all {
			selectID(id, v)
		}
		if updatedAt > 0 {
			out["updated_at"] = updatedAt
		}
		return out
	}
	for _, id := range idList {
		if v, ok := all[id]; ok {
			selectID(id, v)
		}
	}
	return out
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// fetchLiveGeckoPrices pulls CoinGecko on demand for non-USD vs currencies.
func (n *NService) fetchLiveGeckoPrices(ctx context.Context, ids, currency string) (map[string]any, error) {
	idList := splitCSV(ids)
	if len(idList) == 0 {
		idList = constant.COIN_IDS
	}

	cacheKey := fmt.Sprintf("%s:%s:%s", constant.CRYPTO_PRICE_RAW, currency, strings.Join(idList, ","))
	if cached, err := global.NODE_REDIS.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
		var out map[string]any
		if json.Unmarshal([]byte(cached), &out) == nil {
			return out, nil
		}
	}

	keys := strings.Split(global.NODE_CONFIG.Coingecko.ApiKey, ",")
	var apiKeys []string
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k != "" {
			apiKeys = append(apiKeys, k)
		}
	}
	if len(apiKeys) == 0 {
		return nil, errors.New("coingecko api-key is empty")
	}

	query := url.Values{}
	query.Set("ids", strings.Join(idList, ","))
	query.Set("vs_currencies", currency)
	query.Set("include_market_cap", "true")
	query.Set("include_24hr_vol", "true")
	query.Set("include_24hr_change", "true")
	query.Set("include_last_updated_at", "true")
	query.Set("x_cg_demo_api_key", apiKeys[rand.IntN(len(apiKeys))])

	client.URL = constant.CoingeckoGetPrice + "?" + query.Encode()
	var dest map[string]any
	if err := client.HTTPGet(ctx, &dest); err != nil {
		return nil, err
	}

	if b, err := json.Marshal(dest); err == nil {
		_ = global.NODE_REDIS.Set(ctx, cacheKey, string(b), 5*time.Minute).Err()
	}
	return dest, nil
}
