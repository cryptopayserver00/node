package api

import (
	"net/http"
	"node/global"
	"node/model/common"
	"node/model/node/request"
	"node/service"

	"github.com/gin-gonic/gin"
)

func (n *NodeApi) GetSolanaTransactions(c *gin.Context) {
	var res common.Response
	var tx request.GetSolanaTransactions

	err := c.ShouldBind(&tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	result, err := service.NodeService.GetSolanaTransactions(tx)
	if err != nil {
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) GetSolTransactions(c *gin.Context) {
	var res common.Response
	var tx request.GetSolTransactions

	err := c.ShouldBind(&tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	result, err := service.NodeService.GetSolTransactions(tx)
	if err != nil {
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) GetSplTransactions(c *gin.Context) {
	var res common.Response
	var tx request.GetSplTransactions

	err := c.ShouldBind(&tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	result, err := service.NodeService.GetSplTransactions(tx)
	if err != nil {
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", result)

	c.JSON(http.StatusOK, res)
}
