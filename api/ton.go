package api

import (
	"net/http"
	"node/global"
	"node/model/common"
	"node/model/node/request"
	"node/service"

	"github.com/gin-gonic/gin"
)

func (n *NodeApi) GetTonTransactions(c *gin.Context) {
	var res common.Response
	var tx request.GetTonTransactions

	err := c.ShouldBind(&tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	result, err := service.NodeService.GetTonTransactions(tx)
	if err != nil {
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) GetTonCoinTransactions(c *gin.Context) {
	var res common.Response
	var tx request.GetTonCoinTransactions

	err := c.ShouldBind(&tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	result, err := service.NodeService.GetTonCoinTransactions(tx)
	if err != nil {
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) GetTon20Transactions(c *gin.Context) {
	var res common.Response
	var tx request.GetTon20Transactions

	err := c.ShouldBind(&tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	result, err := service.NodeService.GetTon20Transactions(tx)
	if err != nil {
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}
