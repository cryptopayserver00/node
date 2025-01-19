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

func (n *NodeApi) GetTronTransactions(c *gin.Context) {
	var res common.Response
	var trx request.GetTronTransactions

	err := c.ShouldBind(&trx)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rd, _ := json.Marshal(trx)
	global.NODE_LOG.Info("GetTronTransactions: " + string(rd))

	result, err := service.NodeService.GetTronTransactions(trx)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) GetTrxTransactions(c *gin.Context) {
	var res common.Response
	var trx request.GetTrxTransactions

	err := c.ShouldBind(&trx)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rd, _ := json.Marshal(trx)
	global.NODE_LOG.Info("GetTrxTransactions: " + string(rd))

	result, err := service.NodeService.GetTrxTransactions(trx)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) GetTrc20Transactions(c *gin.Context) {
	var res common.Response
	var trc20 request.GetTrc20Transactions

	err := c.ShouldBind(&trc20)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rd, _ := json.Marshal(trc20)
	global.NODE_LOG.Info("GetTrc20Transactions: " + string(rd))

	result, err := service.NodeService.GetTrc20Transactions(trc20)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}
