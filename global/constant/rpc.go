package constant

import (
	"context"
	"math/rand/v2"
	"node/global"
	"node/model/node/request"
	"node/model/node/response"
	"node/utils"
	"strings"
)

// var (
// 	ETHMainnetRPC = []string{
// 		"https://ethereum-rpc.publicnode.com",
// 	}

// 	ETHSepoliaRPC = []string{
// 		"https://ethereum-sepolia.publicnode.com",
// 	}

// 	BSCMainnetRPC = []string{
// 		"https://bsc-dataseed1.binance.org",
// 		"https://bsc-dataseed2.binance.org",
// 		"https://bsc-dataseed3.binance.org",
// 		"https://bsc-dataseed4.binance.org",
// 		"https://bsc-dataseed1.ninicoin.io",
// 		"https://bsc-dataseed2.ninicoin.io",
// 		"https://bsc-dataseed3.ninicoin.io",
// 		"https://bsc-dataseed4.ninicoin.io",
// 	}

// 	BSCTestnetRPC = []string{
// 		"https://data-seed-prebsc-1-s1.binance.org:8545",
// 		"https://data-seed-prebsc-2-s1.binance.org:8545",
// 		"https://data-seed-prebsc-1-s2.binance.org:8545",
// 		"https://data-seed-prebsc-2-s2.binance.org:8545",
// 		"https://data-seed-prebsc-1-s3.binance.org:8545",
// 		"https://data-seed-prebsc-2-s3.binance.org:8545",
// 	}

// 	OPMainnetRPC = []string{
// 		"https://mainnet.optimism.io",
// 		"https://optimism-rpc.publicnode.com",
// 	}

// 	OPSepoliaRPC = []string{
// 		"https://sepolia.optimism.io",
// 		"https://endpoints.omniatech.io/v1/op/sepolia/public",
// 	}

// 	ArbitrumOneRPC = []string{
// 		"https://arb1.arbitrum.io/rpc",
// 		"https://arbitrum-one.publicnode.com",
// 		"https://arbitrum-one-rpc.publicnode.com",
// 	}

// 	ArbitrumNovaRPC = []string{
// 		"https://nova.arbitrum.io/rpc",
// 	}

// 	ArbitrumSepoliaRPC = []string{
// 		"https://sepolia-rollup.arbitrum.io/rpc",
// 	}

// 	SolanaMainnetRPC = []string{
// 		"https://api.mainnet-beta.solana.com",
// 		"https://solana-rpc.publicnode.com",
// 	}

// 	SolanaDevnetRpc = []string{
// 		"https://api.devnet.solana.com",
// 	}

// 	PolMainnetRPC = []string{
// 		"https://polygon-bor-rpc.publicnode.com",
// 		"https://polygon.drpc.org",
// 		"https://poly.api.pocket.network",
// 	}

// 	PolTestnetRPC = []string{
// 		"https://polygon-amoy-bor-rpc.publicnode.com",
// 		"https://rpc-amoy.polygon.technology",
// 	}

// 	AvaxMainnetRPC = []string{
// 		"https://avalanche-c-chain-rpc.publicnode.com",
// 	}

// 	AvaxTestnetRPC = []string{
// 		"https://ava-testnet.public.blastapi.io/ext/bc/C/rpc",
// 		"https://endpoints.omniatech.io/v1/avax/fuji/public",
// 		"https://api.avax-test.network/ext/bc/C/rpc",
// 	}

// 	BaseMainnetRPC = []string{
// 		"https://mainnet.base.org",
// 		"https://developer-access-mainnet.base.org",
// 		"https://base-rpc.publicnode.com",
// 	}

// 	BaseSepoliaRPC = []string{
// 		"https://base-sepolia-rpc.publicnode.com",
// 		"https://sepolia.base.org",
// 		"https://base-sepolia.gateway.tenderly.co",
// 	}

// 	XRPMainnetRPC = []string{
// 		"https://s1.ripple.com:51234",
// 		"https://xrplcluster.com",
// 	}

// 	XRPTestnetRPC = []string{
// 		"https://s.altnet.rippletest.net:51234",
// 	}
// )

func GetAllRPCUrlByChainId(chainId uint) []string {
	switch chainId {
	case ETH_MAINNET:
		return strings.Split(global.NODE_CONFIG.Rpc.EthMainnet, ",")
	case ETH_SEPOLIA:
		return strings.Split(global.NODE_CONFIG.Rpc.EthSepolia, ",")
	case OP_MAINNET:
		return strings.Split(global.NODE_CONFIG.Rpc.OpMainnet, ",")
	case OP_SEPOLIA:
		return strings.Split(global.NODE_CONFIG.Rpc.OpSepolia, ",")
	case BSC_MAINNET:
		return strings.Split(global.NODE_CONFIG.Rpc.BscMainnet, ",")
	case BSC_TESTNET:
		return strings.Split(global.NODE_CONFIG.Rpc.BscTestnet, ",")
	case ARBITRUM_ONE:
		return strings.Split(global.NODE_CONFIG.Rpc.ArbitrumOne, ",")
	case ARBITRUM_NOVA:
		return strings.Split(global.NODE_CONFIG.Rpc.ArbitrumNova, ",")
	case ARBITRUM_SEPOLIA:
		return strings.Split(global.NODE_CONFIG.Rpc.ArbitrumSepolia, ",")
	case POL_MAINNET:
		return strings.Split(global.NODE_CONFIG.Rpc.PolMainnet, ",")
	case POL_TESTNET:
		return strings.Split(global.NODE_CONFIG.Rpc.PolTestnet, ",")
	case AVAX_MAINNET:
		return strings.Split(global.NODE_CONFIG.Rpc.AvaxMainnet, ",")
	case AVAX_TESTNET:
		return strings.Split(global.NODE_CONFIG.Rpc.AvaxTestnet, ",")
	case BASE_MAINNET:
		return strings.Split(global.NODE_CONFIG.Rpc.BaseMainnet, ",")
	case BASE_SEPOLIA:
		return strings.Split(global.NODE_CONFIG.Rpc.BaseSepolia, ",")
	case SOL_MAINNET:
		return strings.Split(global.NODE_CONFIG.Rpc.SolanaMainnet, ",")
	case SOL_DEVNET:
		return strings.Split(global.NODE_CONFIG.Rpc.SolanaDevnet, ",")
	case XRP_MAINNET:
		return strings.Split(global.NODE_CONFIG.Rpc.XrpMainnet, ",")
	case XRP_TESTNET:
		return strings.Split(global.NODE_CONFIG.Rpc.XrpTestnet, ",")
	default:
		panic("not support")
	}
}

// support inner tx(trace_debug)
func GetAlchemyRPCUrlByChainId(chainId uint) string {
	mainnet := IsMainnetSupport(chainId)
	key := GetRandomAlchemyKeyByNetwork(mainnet)

	switch chainId {
	case ETH_MAINNET:
		return "https://eth-mainnet.g.alchemy.com/v2/" + key
	case ETH_SEPOLIA:
		return "https://eth-sepolia.g.alchemy.com/v2/" + key
	case BSC_MAINNET:
		return "https://bnb-mainnet.g.alchemy.com/v2/" + key
	case BSC_TESTNET:
		return "https://bnb-testnet.g.alchemy.com/v2/" + key
	case ARBITRUM_ONE:
		return "https://arb-mainnet.g.alchemy.com/v2/" + key
	case ARBITRUM_NOVA:
		return "https://arbnova-mainnet.g.alchemy.com/v2/" + key
	case ARBITRUM_SEPOLIA:
		return "https://arb-sepolia.g.alchemy.com/v2/" + key
	case OP_MAINNET:
		return "https://opt-mainnet.g.alchemy.com/v2/" + key
	case OP_SEPOLIA:
		return "https://opt-sepolia.g.alchemy.com/v2/" + key
	case SOL_MAINNET:
		return "https://solana-mainnet.g.alchemy.com/v2/" + key
	case SOL_DEVNET:
		return "https://solana-devnet.g.alchemy.com/v2/" + key
	case POL_MAINNET:
		return "https://polygon-mainnet.g.alchemy.com/v2/" + key
	case POL_TESTNET:
		return "https://polygon-amoy.g.alchemy.com/v2/" + key
	case AVAX_MAINNET:
		return "https://avax-mainnet.g.alchemy.com/v2/" + key
	case AVAX_TESTNET:
		return "https://avax-fuji.g.alchemy.com/v2/" + key
	case BASE_MAINNET:
		return "https://base-mainnet.g.alchemy.com/v2/" + key
	case BASE_SEPOLIA:
		return "https://base-sepolia.g.alchemy.com/v2/" + key
	default:
		panic("not support")
	}
}

func GetRealRpcByArray(ctx context.Context, rpcs []string) string {
	for _, rpc := range rpcs {
		client.URL = rpc
		var rpcBlockInfo response.RPCBlockInfo
		var jsonRpcRequest request.JsonRpcRequest
		jsonRpcRequest.Id = 1
		jsonRpcRequest.Jsonrpc = "2.0"
		jsonRpcRequest.Method = "eth_getBlockByNumber"
		jsonRpcRequest.Params = []any{"latest", false}
		err := client.HTTPPost(ctx, jsonRpcRequest, &rpcBlockInfo)
		if err != nil {
			continue
		}

		height, err := utils.HexStringToUint64(rpcBlockInfo.Result.Number)
		if err != nil || !(height > 0) {
			continue
		}
		return rpc
	}
	return ""
}

func GetRPCUrlByChainId(chainId uint) string {
	rpcs := GetAllRPCUrlByChainId(chainId)
	if len(rpcs) == 0 {
		panic("not support")
	}
	return rpcs[rand.IntN(len(rpcs))]
}

func GetRandomAlchemyKeyByNetwork(isMainnet bool) string {
	if isMainnet {
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")[index]
	} else {
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")[index]
	}
}

func GetRandomInnertxAlchemyKeyByNetwork(isMainnet bool) string {
	if isMainnet {
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxMainnetKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxMainnetKey, ",")[index]
	} else {
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxTestnetKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxTestnetKey, ",")[index]
	}
}
