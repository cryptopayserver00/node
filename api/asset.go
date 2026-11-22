package api

import (
	"net/http"
	"node/global"
	"node/model/common"
	"node/model/node/request"
	"node/service"

	"github.com/gin-gonic/gin"
)

func (n *NodeApi) ListAssets(c *gin.Context) {
	var res common.Response
	var req request.ListAssetsRequest
	if err := c.ShouldBind(&req); err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	rows, err := service.NodeService.ListAssets(c.Request.Context(), req)
	if err != nil {
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}
	res = common.OKWithData(rows)
	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) UpsertAsset(c *gin.Context) {
	var res common.Response
	var req request.UpsertAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	asset, err := service.NodeService.UpsertAsset(c.Request.Context(), req)
	if err != nil {
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}
	res = common.OKWithData(asset)
	c.JSON(http.StatusOK, res)
}

func (n *NodeApi) SetAssetEnabled(c *gin.Context) {
	var res common.Response
	var req request.SetAssetEnabledRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.NODE_LOG.Error(err.Error())
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	if err := service.NodeService.SetAssetEnabled(c.Request.Context(), req); err != nil {
		res = common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}
	res = common.OkWithMessage("ok")
	c.JSON(http.StatusOK, res)
}
