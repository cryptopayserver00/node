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

func (n *NodeApi) GetBtcBalance(c *gin.Context) {
	var res common.Response
	var balance request.GetBtcBalance

	err := c.ShouldBind(&balance)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rd, _ := json.Marshal(balance)
	global.NODE_LOG.Info("GetBtcBalance: " + string(rd))

	result, err := service.NodeService.GetBtcBalance(balance)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) GetBtcFeeRate(c *gin.Context) {
	var res common.Response
	var rate request.GetBtcFeeRate

	err := c.ShouldBind(&rate)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rd, _ := json.Marshal(rate)
	global.NODE_LOG.Info("GetBtcFeeRate: " + string(rd))

	result, err := service.NodeService.GetBtcFeeRate(rate)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) GetBtcAddressUtxo(c *gin.Context) {
	var res common.Response
	var utxo request.GetBtcAddressUtxo

	err := c.ShouldBind(&utxo)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rd, _ := json.Marshal(utxo)
	global.NODE_LOG.Info("GetBtcAddressUtxo: " + string(rd))

	result, err := service.NodeService.GetBtcAddressUtxo(utxo)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) PostBtcBroadcast(c *gin.Context) {
	var res common.Response
	var broadcast request.PostBtcBroadcast

	err := c.ShouldBind(&broadcast)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rd, _ := json.Marshal(broadcast)
	global.NODE_LOG.Info("PostBtcBroadcast: " + string(rd))

	result, err := service.NodeService.PostBtcBroadcast(broadcast)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) GetBtcTransactions(c *gin.Context) {
	var res common.Response
	var txs request.GetBtcTransactions

	err := c.ShouldBind(&txs)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rd, _ := json.Marshal(txs)
	global.NODE_LOG.Info("GetBtcTransactions: " + string(rd))

	result, err := service.NodeService.GetBtcTransactions(txs)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) GetBtcTransactionDetail(c *gin.Context) {
	var res common.Response
	var detail request.GetBtcTransactionDetail

	err := c.ShouldBind(&detail)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rd, _ := json.Marshal(detail)
	global.NODE_LOG.Info("GetBtcTransactionDetail: " + string(rd))

	result, err := service.NodeService.GetBtcTransactionDetail(detail)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}
