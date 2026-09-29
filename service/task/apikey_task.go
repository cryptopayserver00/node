package task

import (
	"context"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
	"node/model/node/response/tatum"
	"node/utils"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
)

const ApiKeyTestInterval = 1 * time.Hour

func RunApiKeyTestTask(ctx context.Context) {
	ticker := time.NewTicker(ApiKeyTestInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			RunApiKeyTestCore(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func RunApiKeyTestCore(ctx context.Context) {
	defer utils.HandlePanic()

	global.NODE_LOG.Info("---------- Run Node Testing Task ----------")

	ethNode, ethRate := testLikeEthNodeKey(ctx)
	btcNode, btcRate := testBtcNodeKey(ctx)
	ltcNode, ltcRate := testLtcNodeKey(ctx)
	tronKeys, tronRate := testTronNodeKey(ctx)
	solKeys, solRate := testSolanaNodeKey(ctx)
	tonKeys, tonRate := testTonNodeKey()
	xrpKeys, xrpRate := testXrpNodeKey(ctx)
	bchKeys, bchRate := testBchNodeKey(ctx)

	testAllNode := []string{}
	testAllNode = append(testAllNode, "---------- Run Node Testing Task ----------")
	testAllNode = append(testAllNode, "\n\n")
	testAllNode = append(testAllNode, ethNode...)
	testAllNode = append(testAllNode, "\n")
	testAllNode = append(testAllNode, btcNode...)
	testAllNode = append(testAllNode, "\n")
	testAllNode = append(testAllNode, ltcNode...)
	testAllNode = append(testAllNode, "\n")
	testAllNode = append(testAllNode, tronKeys...)
	testAllNode = append(testAllNode, "\n")
	testAllNode = append(testAllNode, solKeys...)
	testAllNode = append(testAllNode, "\n")
	testAllNode = append(testAllNode, tonKeys...)
	testAllNode = append(testAllNode, "\n")
	testAllNode = append(testAllNode, xrpKeys...)
	testAllNode = append(testAllNode, "\n")
	testAllNode = append(testAllNode, bchKeys...)
	testAllNode = append(testAllNode, fmt.Sprintf("\n\n Total Success Rate: %.2f%%\n", (ethRate+btcRate+ltcRate+tronRate+solRate+tonRate+xrpRate+bchRate)/8*100))
	if len(testAllNode) > 0 {
		go utils.InformToTelegram(strings.Join(testAllNode, ""))
	}
}

func testLikeEthNodeKey(ctx context.Context) ([]string, float64) {
	var failedUrl []string
	failedUrl = append(failedUrl, "Like Ethereum RPC Chain Testing:\n")

	var totalSuccessRate float64

	var nodes []uint
	if global.NODE_CONFIG.Blockchain.SweepMainnet {
		nodes = constant.MainnetEthChain
	} else {
		nodes = constant.TestnetEthChain
	}

	for _, i := range nodes {
		rpcUrl, successRate := testLikeEthByChain(ctx, i)
		failedUrl = append(failedUrl, rpcUrl...)
		totalSuccessRate += successRate
	}

	return failedUrl, totalSuccessRate / float64(len(nodes))
}

func testBtcNodeKey(ctx context.Context) ([]string, float64) {
	var failedUrl []string
	failedUrl = append(failedUrl, "Bitcoin Chain Testing:\n")

	var totalSuccessRate float64

	var nodes []uint
	if global.NODE_CONFIG.Blockchain.SweepMainnet {
		nodes = constant.MainnetBtcChain
	} else {
		nodes = constant.TestnetBtcChain
	}

	for _, i := range nodes {
		url, successRate := testBtcByChain(ctx, i)
		failedUrl = append(failedUrl, url...)
		totalSuccessRate += successRate
	}

	return failedUrl, totalSuccessRate / float64(len(nodes))
}

func testLtcNodeKey(ctx context.Context) ([]string, float64) {
	var failedUrl []string
	failedUrl = append(failedUrl, "Litecoin Chain Testing:\n")

	var totalSuccessRate float64

	var nodes []uint
	if global.NODE_CONFIG.Blockchain.SweepMainnet {
		nodes = constant.MainnetLtcChain
	} else {
		nodes = constant.TestnetLtcChain
	}

	for _, i := range nodes {
		url, successRate := testLtcByChain(ctx, i)
		failedUrl = append(failedUrl, url...)
		totalSuccessRate += successRate
	}

	return failedUrl, totalSuccessRate / float64(len(nodes))
}

func testTronNodeKey(ctx context.Context) ([]string, float64) {
	var failedUrl []string
	failedUrl = append(failedUrl, "Tron Chain Testing:\n")

	var totalSuccessRate float64

	var nodes []uint
	if global.NODE_CONFIG.Blockchain.SweepMainnet {
		nodes = constant.MainnetTronChain
	} else {
		nodes = constant.TestnetTronChain
	}

	for _, i := range nodes {
		url, successRate := testTronByChain(ctx, i)
		failedUrl = append(failedUrl, url...)
		totalSuccessRate += successRate
	}

	return failedUrl, totalSuccessRate / float64(len(nodes))
}

func testSolanaNodeKey(ctx context.Context) ([]string, float64) {
	var failedUrl []string
	failedUrl = append(failedUrl, "Solana Chain Testing:\n")

	var totalSuccessRate float64

	var nodes []uint
	if global.NODE_CONFIG.Blockchain.SweepMainnet {
		nodes = constant.MainnetSolChain
	} else {
		nodes = constant.TestnetSolChain
	}

	for _, i := range nodes {
		rpcUrl, successRate := testSolByChain(ctx, i)
		failedUrl = append(failedUrl, rpcUrl...)
		totalSuccessRate += successRate
	}

	return failedUrl, totalSuccessRate / float64(len(nodes))
}

func testTonNodeKey() ([]string, float64) {
	var failedUrl []string
	failedUrl = append(failedUrl, "Ton Chain Testing:\n")

	var totalSuccessRate float64

	var nodes []uint
	if global.NODE_CONFIG.Blockchain.SweepMainnet {
		nodes = constant.MainnetTonChain
	} else {
		nodes = constant.TestnetTonChain
	}

	for _, i := range nodes {
		rpcUrl, successRate := testTonByChain(i)
		failedUrl = append(failedUrl, rpcUrl...)
		totalSuccessRate += successRate
	}

	return failedUrl, totalSuccessRate / float64(len(nodes))
}

func testXrpNodeKey(ctx context.Context) ([]string, float64) {
	var failedUrl []string
	failedUrl = append(failedUrl, "Xrp Chain Testing:\n")

	var totalSuccessRate float64

	var nodes []uint
	if global.NODE_CONFIG.Blockchain.SweepMainnet {
		nodes = constant.MainnetXrpChain
	} else {
		nodes = constant.TestnetXrpChain
	}

	for _, i := range nodes {
		rpcUrl, successRate := testXrpByChain(ctx, i)
		failedUrl = append(failedUrl, rpcUrl...)
		totalSuccessRate += successRate
	}

	return failedUrl, totalSuccessRate / float64(len(nodes))
}

func testBchNodeKey(ctx context.Context) ([]string, float64) {
	var failedUrl []string
	failedUrl = append(failedUrl, "Bitcoin Cash Chain Testing:\n")

	var totalSuccessRate float64

	var nodes []uint
	if global.NODE_CONFIG.Blockchain.SweepMainnet {
		nodes = constant.MainnetBchChain
	} else {
		nodes = constant.TestnetBchChain
	}

	for _, i := range nodes {
		rpcUrl, successRate := testBchByChain(ctx, i)
		failedUrl = append(failedUrl, rpcUrl...)
		totalSuccessRate += successRate
	}

	return failedUrl, totalSuccessRate / float64(len(nodes))
}

func testLikeEthByChain(ctx context.Context, chainId uint) (status []string, successRate float64) {
	var err error
	allRpc := constant.GetAllRPCUrlByNetwork(chainId)
	var successCount = 0
	if len(allRpc) > 0 {
		for _, v := range allRpc {
			client.URL = v
			var rpcBlockInfo response.RPCBlockInfo
			var jsonRpcRequest request.JsonRpcRequest
			jsonRpcRequest.Id = 1
			jsonRpcRequest.Jsonrpc = "2.0"
			jsonRpcRequest.Method = "eth_getBlockByNumber"
			jsonRpcRequest.Params = []any{"latest", false}
			err = client.HTTPPost(ctx, jsonRpcRequest, &rpcBlockInfo)
			if err != nil {
				global.NODE_LOG.Error(err.Error())
				status = append(status, fmt.Sprintf("❌ | %s -> %s\n", constant.GetChainName(chainId), v))
				continue
			}

			height, err := utils.HexStringToUint64(rpcBlockInfo.Result.Number)
			if err != nil || !(height > 0) {
				global.NODE_LOG.Error(err.Error())
				status = append(status, fmt.Sprintf("❌ | %s -> %s\n", constant.GetChainName(chainId), v))
				continue
			}

			status = append(status, fmt.Sprintf("✅︎ | %s -> %s\n", constant.GetChainName(chainId), v))
			successCount += 1
		}
	}

	return status, float64(successCount) / float64(len(allRpc))
}

func testBtcByChain(ctx context.Context, chainId uint) (status []string, successRate float64) {
	var err error
	var totalNumber int
	var successCount = 0

	allAPiKey := constant.GetAllTatumAPiKey(chainId)
	if len(allAPiKey) > 0 {
		totalNumber += len(allAPiKey)
		for _, v := range allAPiKey {
			client.URL = constant.TatumGetBitcoinInfo
			client.Headers = map[string]string{
				"x-api-key": v,
			}

			var bitcoinInfoResponse tatum.TatumGetBitcoinInfo
			err = client.HTTPGet(ctx, &bitcoinInfoResponse)
			if err != nil {
				global.NODE_LOG.Error(err.Error())
				status = append(status, fmt.Sprintf("❌ | %s -> %s\n", constant.GetChainName(chainId), v))
				continue
			}

			if bitcoinInfoResponse.Blocks > 0 {
				status = append(status, fmt.Sprintf("✅︎ | %s -> %s\n", constant.GetChainName(chainId), v))
				successCount += 1
				continue
			} else {
				status = append(status, fmt.Sprintf("❌ | %s -> %s\n", constant.GetChainName(chainId), v))
				continue
			}
		}
	}

	// mempool
	client.URL = constant.MempoolGetBlockHeightByNetwork(chainId)
	var bitcoinHeight int64
	err = client.HTTPGetUnique(ctx, &bitcoinHeight)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		status = append(status, fmt.Sprintf("❌ | %s -> %s\n", constant.GetChainName(chainId), client.URL))
	}

	if bitcoinHeight > 0 {
		status = append(status, fmt.Sprintf("✅︎ | %s -> %s\n", constant.GetChainName(chainId), client.URL))
		successCount += 1
	} else {
		status = append(status, fmt.Sprintf("❌ | %s -> %s\n", constant.GetChainName(chainId), client.URL))
	}

	totalNumber += 1

	return status, float64(successCount) / float64(totalNumber)
}

func testLtcByChain(ctx context.Context, chainId uint) (status []string, successRate float64) {
	var err error
	var totalNumber int
	var successCount = 0

	allAPiKey := constant.GetAllTatumAPiKey(chainId)
	if len(allAPiKey) > 0 {
		totalNumber += len(allAPiKey)
		for _, v := range allAPiKey {
			client.URL = constant.TatumGetLitecoinInfo
			client.Headers = map[string]string{
				"x-api-key": v,
			}

			var litecoinInfoResponse tatum.TatumGetLitecoinInfo
			err = client.HTTPGet(ctx, &litecoinInfoResponse)
			if err != nil {
				global.NODE_LOG.Error(err.Error())
				status = append(status, fmt.Sprintf("❌ | %s -> %s\n", constant.GetChainName(chainId), v))
				continue
			}

			if litecoinInfoResponse.Blocks > 0 {
				status = append(status, fmt.Sprintf("✅︎ | %s -> %s\n", constant.GetChainName(chainId), v))
				successCount += 1
				continue
			} else {
				status = append(status, fmt.Sprintf("❌ | %s -> %s\n", constant.GetChainName(chainId), v))
				continue
			}
		}
	}

	if global.NODE_CONFIG.Blockchain.SweepMainnet {
		// mempool
		client.URL = constant.MempoolGetBlockHeightByNetwork(chainId)
		var bitcoinHeight int64
		err = client.HTTPGetUnique(ctx, &bitcoinHeight)
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			status = append(status, fmt.Sprintf("❌ | %s -> %s\n", constant.GetChainName(chainId), client.URL))
		}

		if bitcoinHeight > 0 {
			status = append(status, fmt.Sprintf("✅︎ | %s -> %s\n", constant.GetChainName(chainId), client.URL))
			successCount += 1
		} else {
			status = append(status, fmt.Sprintf("❌ | %s -> %s\n", constant.GetChainName(chainId), client.URL))
		}

		totalNumber += 1
	}

	return status, float64(successCount) / float64(totalNumber)
}

func testTronByChain(ctx context.Context, chainId uint) (status []string, successRate float64) {
	var err error
	allAPiKey := constant.GetAllHTTPKeyByNetwork(chainId)
	var successCount = 0
	if len(allAPiKey) > 0 {
		for _, v := range allAPiKey {
			client.URL = constant.TronGetBlockByNetwork(chainId)
			client.Headers = map[string]string{
				"TRON-PRO-API-KEY": v,
			}

			var blockRequest request.TronGetBlockRequest
			blockRequest.Detail = false
			var blockResponse response.TronGetBlockResponse
			err = client.HTTPPost(ctx, blockRequest, &blockResponse)
			if err != nil {
				global.NODE_LOG.Error(err.Error())
				status = append(status, fmt.Sprintf("❌ | %s -> %s\n", constant.GetChainName(chainId), v))
				continue
			}

			if blockResponse.BlockHeader.RawData.Number > 0 {
				status = append(status, fmt.Sprintf("✅︎ | %s -> %s\n", constant.GetChainName(chainId), v))
				successCount += 1
				continue
			} else {
				status = append(status, fmt.Sprintf("❌ | %s -> %s\n", constant.GetChainName(chainId), v))
				continue
			}
		}
	}

	return status, float64(successCount) / float64(len(allAPiKey))
}

func testSolByChain(ctx context.Context, chainId uint) (status []string, successRate float64) {
	var successCount = 0

	allRpc := constant.GetAllRPCUrlByNetwork(chainId)
	if len(allRpc) > 0 {
		for _, v := range allRpc {
			client := rpc.New(v)

			height, err := client.GetSlot(ctx, rpc.CommitmentFinalized)
			if err != nil {
				global.NODE_LOG.Error(err.Error())
				status = append(status, fmt.Sprintf("❌ | %s -> %s\n", constant.GetChainName(chainId), v))
				continue
			}

			if height > 0 {
				status = append(status, fmt.Sprintf("✅︎ | %s -> %s\n", constant.GetChainName(chainId), v))
				successCount += 1
			} else {
				status = append(status, fmt.Sprintf("❌ | %s -> %s\n", constant.GetChainName(chainId), v))
			}
		}
	}

	return status, float64(successCount) / float64(len(allRpc))
}

func testTonByChain(chainId uint) (status []string, successRate float64) {
	return status, 1
}

func testXrpByChain(ctx context.Context, chainId uint) (status []string, successRate float64) {
	var err error
	var successCount = 0

	allRpc := constant.GetAllRPCUrlByNetwork(chainId)
	if len(allRpc) > 0 {
		for _, v := range allRpc {
			client.URL = v
			var rpcBlockDetail response.XrpscanBlockResponse
			var jsonRpcRequest request.XrpJsonRpcRequest
			jsonRpcRequest.Method = "ledger"
			jsonRpcRequest.Params = []map[string]any{
				{
					"transactions": true,
					"expand":       true,
					"api_version":  2,
				},
			}

			err = client.HTTPPost(ctx, jsonRpcRequest, &rpcBlockDetail)
			if err != nil {
				global.NODE_LOG.Error(err.Error())
				status = append(status, fmt.Sprintf("❌ | %s -> %s\n", constant.GetChainName(chainId), v))
				continue
			}

			if rpcBlockDetail.Result.Status == "success" {
				status = append(status, fmt.Sprintf("✅︎ | %s -> %s\n", constant.GetChainName(chainId), v))
				successCount += 1
			} else {
				status = append(status, fmt.Sprintf("❌ | %s -> %s\n", constant.GetChainName(chainId), v))
			}
		}
	}

	return status, float64(successCount) / float64(len(allRpc))
}

func testBchByChain(ctx context.Context, chainId uint) (status []string, successRate float64) {
	var err error
	var totalNumber int
	var successCount = 0

	// mempool
	client.URL = constant.MempoolGetBlockHeightByNetwork(chainId)
	var bitcoinHeight int64
	err = client.HTTPGetUnique(ctx, &bitcoinHeight)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		status = append(status, fmt.Sprintf("❌ | %s -> %s\n", constant.GetChainName(chainId), client.URL))
	}

	if bitcoinHeight > 0 {
		status = append(status, fmt.Sprintf("✅︎ | %s -> %s\n", constant.GetChainName(chainId), client.URL))
		successCount += 1
	} else {
		status = append(status, fmt.Sprintf("❌ | %s -> %s\n", constant.GetChainName(chainId), client.URL))
	}

	totalNumber += 1

	return status, float64(successCount) / float64(totalNumber)
}
