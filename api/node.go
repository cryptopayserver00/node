package api

import (
	"encoding/json"
	"net/http"
	"node/global"
	"node/model/common"
	"node/model/node/request"
	"node/model/node/response"
	"node/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

	rd, _ := json.Marshal(info)
	global.NODE_LOG.Info("GetNetworkInfo: " + string(rd))

	result, err := service.NodeService.GetInfo(info)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
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

	rd, _ := json.Marshal(wallet)
	global.NODE_LOG.Info("StoreWalletAddress: " + string(rd))

	err = service.NodeService.StoreUserWallet(wallet)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
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

	rd, _ := json.Marshal(wallets)
	global.NODE_LOG.Info("BulkStoreUserWallet: " + string(rd))

	result, err := service.NodeService.BulkStorageUserWallets(wallets)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
		res = common.FailWithDetailed(common.Error, err.Error(), result)
		c.JSON(http.StatusOK, res)
		return
	}

	res = common.OKWithData(result)
	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) GetTransactionsByChainAndAddress(c *gin.Context) {
	var res common.Response
	var tx request.TransactionsByChainAndAddress

	err := c.ShouldBind(&tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rd, _ := json.Marshal(tx)
	global.NODE_LOG.Info("GetTransactionsByChainAndAddress: " + string(rd))

	result, total, err := service.NodeService.GetTransactionsByChainAndAddress(tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error(), zap.Error(err))
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
