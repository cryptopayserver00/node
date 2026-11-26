package service

import (
	"fmt"
	"node/global"
	"node/model/node/request"
	"node/model/node/response"

	"github.com/gin-gonic/gin"
)

func (n *NService) GetBscTransactions(c *gin.Context, req request.GetBscTransactions) ([]response.ClientTransaction, error) {
	ctx := c.Request.Context()
	txs, err := n.GetEvmTransactionsByExplorer(ctx, req.ChainId, req.Address, req.Asset)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("bsc explorer txs failed: %v", err))
		return []response.ClientTransaction{}, nil
	}
	return txs, nil
}
