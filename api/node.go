package api

import (
	"net/http"
	"node/global"
	"node/model/common"
	"node/model/node/request"
	"node/model/node/response"
	"node/service"

	"github.com/gin-gonic/gin"
)

func (n *NodeApi) GetNetworkInfo(c *gin.Context) {
	var res common.Response
	var info request.GetNetworkInfo

	err := c.ShouldBind(&info)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	result, err := service.NodeService.GetInfo(c, info)
	if err != nil {
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OKWithData(result)
	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) StoreWalletAddress(c *gin.Context) {
	var res common.Response
	var wallet request.StoreUserWallet

	err := c.ShouldBind(&wallet)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	err = service.NodeService.StoreUserWallet(c, wallet)
	if err != nil {
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithMessage("store successfully")
	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) BulkStoreUserWallet(c *gin.Context) {
	var res common.Response
	var wallets request.BulkStoreUserWallet

	err := c.ShouldBind(&wallets)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	result, err := service.NodeService.BulkStorageUserWallets(c, wallets)
	if err != nil {
		res = common.FailWithDetailed(common.Error, err.Error(), result)
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OKWithData(result)
	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) GetTransactionsByChainAndAddress(c *gin.Context) {
	ctx := c.Request.Context()

	var res common.Response
	var tx request.TransactionsByChainAndAddress

	err := c.ShouldBind(&tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	result, total, err := service.NodeService.GetTransactionsByChainAndAddress(ctx, tx)
	if err != nil {
		res = common.FailWithDetailed(common.Error, err.Error(), result)
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OkWithDetailed(common.Success, "Request data successful", response.OwnListResponse{
		Transactions: result,
		Total:        total,
		Page:         tx.Page,
		PageSize:     tx.PageSize,
	})

	c.JSON(http.StatusOK, res)
}
