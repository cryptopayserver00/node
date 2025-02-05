package service

import (
	"errors"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
	"node/sweep/utils"
	"node/utils/wallet"

	"github.com/gin-gonic/gin"
)

func (m *NService) GetFreeCoin(c *gin.Context, req request.GetFreeCoin) (freeCoin response.FreeCoinResponse, err error) {

	var hash string

	if !utils.IsFreeCoinSupport(req.ChainId, req.Coin) {
		return freeCoin, errors.New("do not support the coin for the free coin")
	}

	if !constant.IsAddressSupport(req.ChainId, req.Address) {
		return freeCoin, errors.New("do not support the address")
	}

	hash, err = wallet.TransferFreeCoinToReceiveAddress(req.ChainId, req.Coin, req.Address, req.Amount)
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
