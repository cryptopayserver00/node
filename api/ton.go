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

	rd, _ := json.Marshal(tx)
	global.NODE_LOG.Info("GetTonTransactions: " + string(rd))

	result, err := service.NodeService.GetTonTransactions(tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
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

	rd, _ := json.Marshal(tx)
	global.NODE_LOG.Info("GetTonCoinTransactions: " + string(rd))

	result, err := service.NodeService.GetTonCoinTransactions(tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
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

	rd, _ := json.Marshal(tx)
	global.NODE_LOG.Info("GetTon20Transactions: " + string(rd))

	result, err := service.NodeService.GetTon20Transactions(tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}
