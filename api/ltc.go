package api

import (
	"encoding/json"
	"net/http"
	"node/global"
	"node/model/common"
	"node/model/node/request"
	"node/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (n *NodeApi) GetLtcBalance(c *gin.Context) {
	var res common.Response
	var balance request.GetLtcBalance

	err := c.ShouldBind(&balance)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rd, _ := json.Marshal(balance)
	global.NODE_LOG.Info("GetLtcBalance: " + string(rd))

	result, err := service.NodeService.GetLtcBalance(balance)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) GetLtcFeeRate(c *gin.Context) {
	var res common.Response
	var rate request.GetLtcFeeRate

	err := c.ShouldBind(&rate)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rd, _ := json.Marshal(rate)
	global.NODE_LOG.Info("GetLtcFeeRate: " + string(rd))

	result, err := service.NodeService.GetLtcFeeRate(rate)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) PostLtcBroadcast(c *gin.Context) {
	var res common.Response
	var broadcast request.PostLtcBroadcast

	err := c.ShouldBind(&broadcast)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rd, _ := json.Marshal(broadcast)
	global.NODE_LOG.Info("PostLtcBroadcast: " + string(rd))

	result, err := service.NodeService.PostLtcBroadcast(broadcast)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) GetLtcTransactions(c *gin.Context) {
	var res common.Response
	var broadcast request.GetLtcTransactions

	err := c.ShouldBind(&broadcast)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rd, _ := json.Marshal(broadcast)
	global.NODE_LOG.Info("GetLtcTransactions: " + string(rd))

	result, err := service.NodeService.GetLtcTransactions(broadcast)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) GetLtcTxByHash(c *gin.Context) {
	var res common.Response
	var tx request.GetLtcTxByHash

	err := c.ShouldBind(&tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rd, _ := json.Marshal(tx)
	global.NODE_LOG.Info("GetLtcTxByHash: " + string(rd))

	result, err := service.NodeService.GetLtcTxByHash(tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) GetLtcAddressUtxo(c *gin.Context) {
	var res common.Response
	var utxo request.GetLtcAddressUtxo

	err := c.ShouldBind(&utxo)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rd, _ := json.Marshal(utxo)
	global.NODE_LOG.Info("GetLtcAddressUtxo: " + string(rd))

	result, err := service.NodeService.GetLtcAddressUtxo(utxo)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}
