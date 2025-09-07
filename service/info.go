package service

import (
	"errors"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
	"node/sweep/setup"

	"github.com/gin-gonic/gin"
)

func (n *NService) GetInfo(c *gin.Context, req request.GetNetworkInfo) (result response.NetworkInfo, err error) {
	var networkResponse response.NetworkInfo

	if !constant.IsNetworkSupport(req.ChainId) {
		return networkResponse, errors.New("do not support the network")
	}

	networkResponse.ChainId = req.ChainId

	latest, cache, sweep, err := setup.GetBlockHeight(c, req.ChainId)
	if err != nil {
		return networkResponse, err
	}

	networkResponse.LatestBlock = latest
	networkResponse.CacheBlock = cache
	networkResponse.SweepBlock = sweep

	return networkResponse, nil
}
