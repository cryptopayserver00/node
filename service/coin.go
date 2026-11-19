package service

import (
	"errors"
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

	if strings.TrimSpace(fc.MaxAmount) == "" {
		return freeCoin, errors.New("freecoin max-amount is not configured")
	}

	if strings.TrimSpace(req.Amount) == "" {
		return freeCoin, errors.New("amount is required")
	}

	if err = validateFreeCoinAmount(req.Amount, fc.MaxAmount); err != nil {
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

func validateFreeCoinAmount(amount, maxAmount string) error {
	a, ok := new(big.Rat).SetString(strings.TrimSpace(amount))
	if !ok {
		return errors.New("invalid amount")
	}
	if a.Sign() <= 0 {
		return errors.New("amount must be positive")
	}

	max, ok := new(big.Rat).SetString(strings.TrimSpace(maxAmount))
	if !ok || max.Sign() <= 0 {
		return errors.New("invalid freecoin max-amount")
	}

	if a.Cmp(max) > 0 {
		return errors.New("amount exceeds freecoin max-amount")
	}
	return nil
}
