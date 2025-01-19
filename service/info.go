package service

import (
	"errors"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
)

func (n *NService) GetInfo(info request.GetNetworkInfo) (networkInfo response.NetworkInfo, err error) {
	if !constant.IsNetworkSupport(info.ChainId) {
		return networkInfo, errors.New("do not support the network")
	}

	networkInfo.ChainId = info.ChainId

	if constant.IsNetworkSupportTatum(info.ChainId) {
		networkInfo.TatumUrl = constant.TatumAPI
		networkInfo.TatumKey = constant.GetTatumRandomKeyByNetwork(info.ChainId)
	}

	networkInfo.RPCUrl = constant.GetRPCUrlByNetwork((info.ChainId))
	networkInfo.HTTPUrl = constant.GetHttpUrlByNetwork(info.ChainId)
	networkInfo.HTTPKey = constant.GetRandomHTTPKeyByNetwork(info.ChainId)

	return
}
