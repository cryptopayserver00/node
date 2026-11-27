package service

import (
	"node/model/node/request"
	"node/model/node/response"

	"github.com/gin-gonic/gin"
)

func (n *NService) GetBscTransactions(c *gin.Context, req request.GetBscTransactions) ([]response.ClientTransaction, error) {
	return n.GetEvmTransactions(c.Request.Context(), req)
}
