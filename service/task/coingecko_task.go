package task

import (
	"context"
	"encoding/json"
	"math/rand/v2"
	"net/url"
	"node/global"
	"node/global/constant"
	"node/model/node/response"
	"node/utils"
	"strings"
	"time"
)

const CoingeckoDuration = 300 // seconds

func RunCoingeckoTask(ctx context.Context) {
	RunCoingeckoCore(ctx)
	for {
		if !waitUntilNextInterval(ctx, CoingeckoDuration) {
			return
		}
		RunCoingeckoCore(ctx)
	}
}

func RunCoingeckoCore(ctx context.Context) {
	defer utils.HandlePanic()

	global.NODE_LOG.Info("---------- Run Coingecko Task ----------")

	if global.NODE_REDIS == nil {
		global.NODE_LOG.Error("redis not initialized, skip coingecko task")
		return
	}

	keys := strings.Split(global.NODE_CONFIG.Coingecko.ApiKey, ",")
	keys = filterNonEmpty(keys)
	if len(keys) == 0 {
		global.NODE_LOG.Error("coingecko api-key is empty")
		return
	}
	key := keys[rand.IntN(len(keys))]

	query := url.Values{}
	query.Set("ids", strings.Join(constant.COIN_IDS, ","))
	query.Set("vs_currencies", "USD")
	query.Set("include_market_cap", "true")
	query.Set("include_24hr_vol", "true")
	query.Set("include_24hr_change", "true")
	query.Set("include_last_updated_at", "true")
	query.Set("x_cg_demo_api_key", key)

	reqURL := constant.CoingeckoGetPrice + "?" + query.Encode()

	client.URL = reqURL
	var cryptoResponse response.CoingeckoPrice
	if err := client.HTTPGet(ctx, &cryptoResponse); err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	custom := mapToSymbolPrice(cryptoResponse)

	rawByte, err := json.Marshal(cryptoResponse)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}
	symbolByte, err := json.Marshal(custom)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}
	metaByte, err := json.Marshal(response.CryptoPriceMeta{UpdatedAt: time.Now().UnixMilli()})
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pipe := global.NODE_REDIS.Pipeline()
	pipe.Set(ctxTimeout, constant.CRYPTO_PRICE_RAW, string(rawByte), time.Hour)
	pipe.Set(ctxTimeout, constant.CRYPTO_PRICE_SYMBOL, string(symbolByte), time.Hour)
	pipe.Set(ctxTimeout, constant.CRYPTO_PRICE_META, string(metaByte), time.Hour)
	if _, err := pipe.Exec(ctxTimeout); err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}
}

func mapToSymbolPrice(src response.CoingeckoPrice) response.CustomCoingeckoPrice {
	return response.CustomCoingeckoPrice{
		BTC: src.BTC, ETH: src.ETH, USDT: src.USDT, BNB: src.BNB, SOL: src.SOL,
		USDC: src.USDC, XRP: src.XRP, TON: src.TON, DOGE: src.DOGE, ADA: src.ADA,
		TRX: src.TRX, AVAX: src.AVAX, SHIB: src.SHIB, DOT: src.DOT, LINK: src.LINK,
		BCH: src.BCH, DAI: src.DAI, LTC: src.LTC, POL: src.POL, UNI: src.UNI,
		PEPE: src.PEPE, FIL: src.FIL, ARB: src.ARB, OP: src.OP, WIF: src.WIF,
		NOT: src.NOT, BONK: src.BONK, AAVE: src.AAVE, FLOKI: src.FLOKI, JUP: src.JUP,
		ENS: src.ENS, W: src.W, MEW: src.MEW, WBTC: src.WBTC, WETH: src.WETH,
		WSOL: src.WSOL, SAND: src.SAND,
	}
}

func filterNonEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
