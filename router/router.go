package router

import (
	"node/api"

	"github.com/gin-gonic/gin"
)

type MainRouter struct{}

func (mRouter *MainRouter) InitRouter(Router *gin.RouterGroup) {
	router := Router.Group("/")
	api := new(api.NodeApi)
	{
		router.GET("test", api.Test)
		router.GET("networkInfo", api.GetNetworkInfo)
		router.POST("storeWalletAddress", api.StoreWalletAddress)
		router.POST("bulkStoreUserWallet", api.BulkStoreUserWallet)
		// router.POST("storeChainContract", api.StoreChainContract)
		// router.POST("bulkStoreChainContract", api.BulkStoreChainContract)

		// router.GET("getTransactionByChainAndHash", api.GetTransactionByChainAndHash)
		router.GET("getTransactionsByChainAndAddress", api.GetTransactionsByChainAndAddress)

		router.GET("coinFree", api.GetFreeCoin)
	}

	// router.Use(middleware.Wss())
	// {
	// 	// websocket
	// 	router.GET("ws", api.WsForTxInfo)
	// }

	node := Router.Group("node")
	{
		node.GET("networkInfo", api.GetNetworkInfo)
	}

	nodeForEth := node.Group("eth")
	{
		nodeForEth.GET("getTransactions", api.GetEthTransactions)
		nodeForEth.GET("getPendingTransaction", api.GetEthPendingTransaction)
	}

	nodeForBsc := node.Group("bsc")
	{
		nodeForBsc.GET("getTransactions", api.GetBscTransactions)
	}

	nodeForBtc := node.Group("btc")
	{
		nodeForBtc.GET("getBalance", api.GetBtcBalance)
		nodeForBtc.GET("getFeeRate", api.GetBtcFeeRate)
		nodeForBtc.GET("getAddressUtxo", api.GetBtcAddressUtxo)
		nodeForBtc.POST("postBroadcast", api.PostBtcBroadcast)
		nodeForBtc.GET("getTransactions", api.GetBtcTransactions)
		nodeForBtc.GET("getTransactionDetail", api.GetBtcTransactionDetail)
	}

	nodeForLtc := node.Group("ltc")
	{
		nodeForLtc.GET("getBalance", api.GetLtcBalance)
		nodeForLtc.GET("getFeeRate", api.GetLtcFeeRate)
		nodeForLtc.POST("postBroadcast", api.PostLtcBroadcast)
		nodeForLtc.GET("getTransactions", api.GetLtcTransactions)
		nodeForLtc.GET("getTxByHash", api.GetLtcTxByHash)
		nodeForLtc.GET("getAddressUtxo", api.GetLtcAddressUtxo)
	}

	nodeForTron := node.Group("tron")
	{
		nodeForTron.GET("getTransactions", api.GetTronTransactions)
		nodeForTron.GET("getTrxTransactions", api.GetTrxTransactions)
		nodeForTron.GET("getTrc20Transactions", api.GetTrc20Transactions)
	}

	nodeForSolana := node.Group("solana")
	{
		nodeForSolana.GET("getTransactions", api.GetSolanaTransactions)
		nodeForSolana.GET("getSolTransactions", api.GetSolTransactions)
		nodeForSolana.GET("getSplTransactions", api.GetSplTransactions)
	}

	nodeForTon := node.Group("ton")
	{
		nodeForTon.GET("getTransactions", api.GetTonTransactions)
		nodeForTon.GET("getTonTransactions", api.GetTonCoinTransactions)
		nodeForTon.GET("getTon20Transactions", api.GetTon20Transactions)
	}

	nodeForXrp := node.Group("xrp")
	{
		nodeForXrp.GET("getTransactions", api.GetXrpTransactions)
	}

	nodeForBch := node.Group("bch")
	{
		nodeForBch.GET("getTransactions", api.GetBchTransactions)
	}

	nodeForArb := node.Group("arb")
	{
		nodeForArb.GET("getTransactions", api.GetArbTransactions)
	}

	nodeForAvax := node.Group("avax")
	{
		nodeForAvax.GET("getTransactions", api.GetAvaxTransactions)
	}

	nodeForPol := node.Group("pol")
	{
		nodeForPol.GET("getTransactions", api.GetPolTransactions)
	}

	nodeForBase := node.Group("base")
	{
		nodeForBase.GET("getTransactions", api.GetBaseTransactions)
	}

	nodeForOp := node.Group("op")
	{
		nodeForOp.GET("getTransactions", api.GetOpTransactions)
	}
}
