package api

import (
	"net/http"
	"node/global"
	"node/model/common"
	"node/model/node/request"
	"node/service"

	"github.com/gin-gonic/gin"
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

	result, err := service.NodeService.GetBtcBalance(c, balance)
	if err != nil {
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

	result, err := service.NodeService.GetBtcFeeRate(c, rate)
	if err != nil {
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

	result, err := service.NodeService.GetBtcAddressUtxo(c, utxo)
	if err != nil {
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

	result, err := service.NodeService.PostBtcBroadcast(c, broadcast)
	if err != nil {
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

	result, err := service.NodeService.GetBtcTransactions(c, txs)
	if err != nil {
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

	result, err := service.NodeService.GetBtcTransactionDetail(c, detail)
	if err != nil {
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}
