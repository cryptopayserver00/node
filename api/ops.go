package api

import (
	"net/http"
	"node/global"
	"node/model/common"
	"strconv"

	"node/service"

	"github.com/gin-gonic/gin"
)

func (n *NodeApi) GetOpsStatus(c *gin.Context) {
	includeFailed := c.Query("include_failed") == "1" || c.Query("include_failed") == "true"
	status, err := service.NodeService.GetOpsStatus(c.Request.Context(), includeFailed)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		res := common.FailWithMessage(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}
	c.JSON(http.StatusOK, common.OKWithData(status))
}

func (n *NodeApi) RetryFailedOutbox(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		idStr = c.PostForm("id")
	}
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusOK, common.FailWithMessage("invalid id"))
		return
	}
	if err := service.NodeService.RetryFailedOutbox(c.Request.Context(), uint(id64)); err != nil {
		c.JSON(http.StatusOK, common.FailWithMessage(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OkWithMessage("queued for retry"))
}
