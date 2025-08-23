package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (n *NodeApi) Test(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}
