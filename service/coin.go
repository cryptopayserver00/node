package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
	"node/sweep/utils"
	"node/utils/wallet"
	"strings"

	"github.com/gin-gonic/gin"
)

func (m *NService) GetFreeCoin(c *gin.Context, req request.GetFreeCoin) (freeCoin response.FreeCoinResponse, err error) {
	ctx := c.Request.Context()
	fc := global.NODE_CONFIG.FreeCoin

	if !fc.Enabled {
		return freeCoin, errors.New("freecoin is disabled")
	}

	if strings.TrimSpace(req.Amount) == "" {
		return freeCoin, errors.New("amount is required")
	}

	maxCoin, err := resolveFreeCoinMaxAmount(ctx, req.Coin, fc.MaxUsd, fc.MaxAmount)
	if err != nil {
		return freeCoin, err
	}

	if err = validateFreeCoinAmount(req.Amount, maxCoin); err != nil {
		return freeCoin, err
	}

	if !utils.IsFreeCoinSupport(req.ChainId, req.Coin) {
		return freeCoin, errors.New("do not support the coin for the free coin")
	}

	if !constant.IsAddressSupport(req.ChainId, req.Address) {
		return freeCoin, errors.New("do not support the address")
	}

	hash, err := wallet.TransferFreeCoinToReceiveAddress(ctx, req.ChainId, req.Coin, req.Address, req.Amount)
	if err != nil {
		return
	}

	if hash == "" {
		return freeCoin, errors.New("the transaction fails, please try it again")
	}

	freeCoin.ChainId = req.ChainId
	freeCoin.Hash = hash

	return freeCoin, nil
}

func validateFreeCoinAmount(amount string, maxAmount *big.Rat) error {
	a, ok := new(big.Rat).SetString(strings.TrimSpace(amount))
	if !ok {
		return errors.New("invalid amount")
	}
	if a.Sign() <= 0 {
		return errors.New("amount must be positive")
	}
	if maxAmount == nil || maxAmount.Sign() <= 0 {
		return errors.New("invalid freecoin max-amount")
	}
	if a.Cmp(maxAmount) > 0 {
		return fmt.Errorf("amount exceeds freecoin max (~%s)", maxAmount.FloatString(8))
	}
	return nil
}

// resolveFreeCoinMaxAmount converts max-usd into a coin-unit ceiling using market price.
// Prefer max-usd; fall back to absolute max-amount for older configs.
func resolveFreeCoinMaxAmount(ctx context.Context, coin, maxUsd, maxAmount string) (*big.Rat, error) {
	maxUsd = strings.TrimSpace(maxUsd)
	maxAmount = strings.TrimSpace(maxAmount)

	if maxUsd != "" {
		usd, ok := new(big.Rat).SetString(maxUsd)
		if !ok || usd.Sign() <= 0 {
			return nil, errors.New("invalid freecoin max-usd")
		}
		price, err := getFreeCoinUsdPrice(ctx, coin)
		if err != nil {
			return nil, err
		}
		if price <= 0 {
			return nil, fmt.Errorf("invalid usd price for %s", coin)
		}
		priceRat := new(big.Rat).SetFloat64(price)
		if priceRat.Sign() <= 0 {
			return nil, fmt.Errorf("invalid usd price for %s", coin)
		}
		return new(big.Rat).Quo(usd, priceRat), nil
	}

	if maxAmount != "" {
		max, ok := new(big.Rat).SetString(maxAmount)
		if !ok || max.Sign() <= 0 {
			return nil, errors.New("invalid freecoin max-amount")
		}
		return max, nil
	}

	return nil, errors.New("freecoin max-usd is not configured")
}

func getFreeCoinUsdPrice(ctx context.Context, coin string) (float64, error) {
	sym := normalizeFreeCoinSymbol(coin)

	if isStableFreeCoin(sym) {
		return 1, nil
	}

	if global.NODE_REDIS != nil {
		data, err := global.NODE_REDIS.Get(ctx, constant.CRYPTO_PRICE_SYMBOL).Result()
		if err == nil && data != "" {
			var prices map[string]struct {
				USD float64 `json:"usd"`
			}
			if json.Unmarshal([]byte(data), &prices) == nil {
				if p, ok := prices[strings.ToLower(sym)]; ok && p.USD > 0 {
					return p.USD, nil
				}
			}
		}
	}

	if p, ok := freeCoinFallbackUsd[strings.ToUpper(sym)]; ok && p > 0 {
		return p, nil
	}

	return 0, fmt.Errorf("no usd price for %s", coin)
}

func normalizeFreeCoinSymbol(coin string) string {
	sym := strings.ToUpper(strings.TrimSpace(coin))
	switch sym {
	case "WETH":
		return "ETH"
	case "WBTC":
		return "BTC"
	case "WSOL":
		return "SOL"
	case "MATIC":
		return "POL"
	default:
		return sym
	}
}

func isStableFreeCoin(sym string) bool {
	switch strings.ToUpper(sym) {
	case "USDT", "USDC", "DAI", "BUSD", "FDUSD", "TUSD", "USDD":
		return true
	default:
		return false
	}
}

// Rough mainnet USD anchors used only when CoinGecko cache is unavailable.
var freeCoinFallbackUsd = map[string]float64{
	"BTC":  95000,
	"ETH":  3500,
	"BNB":  600,
	"SOL":  150,
	"XRP":  2.2,
	"TON":  5,
	"DOGE": 0.2,
	"ADA":  0.7,
	"TRX":  0.25,
	"AVAX": 35,
	"SHIB": 0.00002,
	"DOT":  7,
	"LINK": 15,
	"BCH":  450,
	"LTC":  100,
	"POL":  0.4,
	"UNI":  10,
	"PEPE": 0.00001,
	"FIL":  5,
	"ARB":  0.8,
	"OP":   1.5,
	"AAVE": 180,
	"SAND": 0.5,
}
