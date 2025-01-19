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

func (n *NodeApi) GetEthTransactions(c *gin.Context) {
	var res common.Response
	var tx request.GetEthTransactions

	err := c.ShouldBind(&tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rd, _ := json.Marshal(tx)
	global.NODE_LOG.Info("GetEthTransactions: " + string(rd))

	result, err := service.NodeService.GetEthTransactions(tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithDetailed(common.Error, err.Error(), result)
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) GetEthPendingTransaction(c *gin.Context) {
	var res common.Response
	var tx request.GetEthPendingTransaction

	err := c.ShouldBind(&tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rd, _ := json.Marshal(tx)
	global.NODE_LOG.Info("GetEthPendingTransaction: " + string(rd))

	result, err := service.NodeService.GetEthPendingTransaction(tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithDetailed(common.Error, err.Error(), result)
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}