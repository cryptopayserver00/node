package router

import (
	"node/api"
	"node/middleware"

	"github.com/gin-gonic/gin"
)

type MainRouter struct{}

func (mRouter *MainRouter) InitRouter(Router *gin.RouterGroup) {
	apiHandler := new(api.NodeApi)

	// 公开探测；业务接口一律走鉴权 + IP 限流
	Router.GET("test", apiHandler.Test)

	protected := Router.Group("/")
	protected.Use(middleware.APIAuth(), middleware.IPRateLimit())
	{
		protected.GET("networkInfo", apiHandler.GetNetworkInfo)
		protected.POST("storeWalletAddress", apiHandler.StoreWalletAddress)
		protected.POST("bulkStoreUserWallet", apiHandler.BulkStoreUserWallet)
		protected.GET("getTransactionsByChainAndAddress", apiHandler.GetTransactionsByChainAndAddress)
		protected.GET("coinFree", apiHandler.GetFreeCoin)
		protected.GET("listAssets", apiHandler.ListAssets)
		protected.POST("upsertAsset", apiHandler.UpsertAsset)
		protected.POST("setAssetEnabled", apiHandler.SetAssetEnabled)
		protected.GET("opsStatus", apiHandler.GetOpsStatus)
		protected.POST("retryFailedOutbox", apiHandler.RetryFailedOutbox)
	}

	ws := Router.Group("/")
	ws.Use(middleware.DeshopWss())
	{
		ws.GET("deshopWs", apiHandler.WsForTxInfo)
	}

	node := Router.Group("node")
	node.Use(middleware.APIAuth(), middleware.IPRateLimit())
	{
		node.GET("networkInfo", apiHandler.GetNetworkInfo)
	}

	nodeForEth := node.Group("eth")
	{
		nodeForEth.GET("getTransactions", apiHandler.GetEthTransactions)
		nodeForEth.GET("getPendingTransaction", apiHandler.GetEthPendingTransaction)
	}

	nodeForBsc := node.Group("bsc")
	{
		nodeForBsc.GET("getTransactions", apiHandler.GetBscTransactions)
	}

	nodeForBtc := node.Group("btc")
	{
		nodeForBtc.GET("getBalance", apiHandler.GetBtcBalance)
		nodeForBtc.GET("getFeeRate", apiHandler.GetBtcFeeRate)
		nodeForBtc.GET("getAddressUtxo", apiHandler.GetBtcAddressUtxo)
		nodeForBtc.POST("postBroadcast", apiHandler.PostBtcBroadcast)
		nodeForBtc.GET("getTransactions", apiHandler.GetBtcTransactions)
		nodeForBtc.GET("getTransactionDetail", apiHandler.GetBtcTransactionDetail)
	}

	nodeForLtc := node.Group("ltc")
	{
		nodeForLtc.GET("getBalance", apiHandler.GetLtcBalance)
		nodeForLtc.GET("getFeeRate", apiHandler.GetLtcFeeRate)
		nodeForLtc.POST("postBroadcast", apiHandler.PostLtcBroadcast)
		nodeForLtc.GET("getTransactions", apiHandler.GetLtcTransactions)
		nodeForLtc.GET("getTxByHash", apiHandler.GetLtcTxByHash)
		nodeForLtc.GET("getAddressUtxo", apiHandler.GetLtcAddressUtxo)
	}

	nodeForTron := node.Group("tron")
	{
		nodeForTron.GET("getTransactions", apiHandler.GetTronTransactions)
		nodeForTron.GET("getTrxTransactions", apiHandler.GetTrxTransactions)
		nodeForTron.GET("getTrc20Transactions", apiHandler.GetTrc20Transactions)
	}

	nodeForSolana := node.Group("solana")
	{
		nodeForSolana.GET("getTransactions", apiHandler.GetSolanaTransactions)
		nodeForSolana.GET("getSolTransactions", apiHandler.GetSolTransactions)
		nodeForSolana.GET("getSplTransactions", apiHandler.GetSplTransactions)
	}

	nodeForTon := node.Group("ton")
	{
		nodeForTon.GET("getTransactions", apiHandler.GetTonTransactions)
		nodeForTon.GET("getTonTransactions", apiHandler.GetTonCoinTransactions)
		nodeForTon.GET("getTon20Transactions", apiHandler.GetTon20Transactions)
	}

	nodeForXrp := node.Group("xrp")
	{
		nodeForXrp.GET("getTransactions", apiHandler.GetXrpTransactions)
	}

	nodeForBch := node.Group("bch")
	{
		nodeForBch.GET("getTransactions", apiHandler.GetBchTransactions)
	}

	nodeForArb := node.Group("arb")
	{
		nodeForArb.GET("getTransactions", apiHandler.GetArbTransactions)
	}

	nodeForArbNova := node.Group("arbnova")
	{
		nodeForArbNova.GET("getTransactions", apiHandler.GetArbNovaTransactions)
	}

	nodeForAvax := node.Group("avax")
	{
		nodeForAvax.GET("getTransactions", apiHandler.GetAvaxTransactions)
	}

	nodeForPol := node.Group("pol")
	{
		nodeForPol.GET("getTransactions", apiHandler.GetPolTransactions)
	}

	nodeForBase := node.Group("base")
	{
		nodeForBase.GET("getTransactions", apiHandler.GetBaseTransactions)
	}

	nodeForOp := node.Group("op")
	{
		nodeForOp.GET("getTransactions", apiHandler.GetOpTransactions)
	}
}
