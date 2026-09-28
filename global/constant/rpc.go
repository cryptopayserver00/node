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

var (
	ETHGeneralMainnetRPC = []string{
		"https://ethereum-rpc.publicnode.com",
	}

	ETHGeneralSepoliaRPC = []string{
		"https://ethereum-sepolia.publicnode.com",
	}

	ETHMainnetRPC = []string{
		"https://ethereum-rpc.publicnode.com",
	}

	ETHSepoliaRPC = []string{
		"https://ethereum-sepolia.publicnode.com",
	}

	BSCMainnetRPC = []string{
		"https://bsc-dataseed1.binance.org",
		"https://bsc-dataseed2.binance.org",
		"https://bsc-dataseed3.binance.org",
		"https://bsc-dataseed4.binance.org",
		"https://bsc-dataseed1.ninicoin.io",
		"https://bsc-dataseed2.ninicoin.io",
		"https://bsc-dataseed3.ninicoin.io",
		"https://bsc-dataseed4.ninicoin.io",
	}

	BSCTestnetRPC = []string{
		"https://data-seed-prebsc-1-s1.binance.org:8545",
		"https://data-seed-prebsc-2-s1.binance.org:8545",
		"https://data-seed-prebsc-1-s2.binance.org:8545",
		"https://data-seed-prebsc-2-s2.binance.org:8545",
		"https://data-seed-prebsc-1-s3.binance.org:8545",
		"https://data-seed-prebsc-2-s3.binance.org:8545",
	}

	OPMainnetRPC = []string{
		"https://mainnet.optimism.io",
		"https://optimism-rpc.publicnode.com",
	}

	OPSepoliaRPC = []string{
		"https://sepolia.optimism.io",
		"https://endpoints.omniatech.io/v1/op/sepolia/public",
	}

	ArbitrumOneRPC = []string{
		"https://arb1.arbitrum.io/rpc",
		"https://arbitrum-one.publicnode.com",
		"https://arbitrum-one-rpc.publicnode.com",
	}

	ArbitrumNovaRPC = []string{
		"https://nova.arbitrum.io/rpc",
	}

	ArbitrumSepoliaRPC = []string{
		"https://sepolia-rollup.arbitrum.io/rpc",
	}

	SolanaMainnetRPC = []string{
		"https://api.mainnet-beta.solana.com",
		"https://solana-rpc.publicnode.com",
	}

	SolanaDevnetRpc = []string{
		"https://api.devnet.solana.com",
	}

	PolMainnetRPC = []string{
		"https://polygon-bor-rpc.publicnode.com",
	}

	PolTestnetRPC = []string{
		"https://polygon-amoy-bor-rpc.publicnode.com",
		"https://rpc-amoy.polygon.technology",
	}

	AvaxMainnetRPC = []string{
		"https://avalanche-c-chain-rpc.publicnode.com",
	}

	AvaxTestnetRPC = []string{
		"https://ava-testnet.public.blastapi.io/ext/bc/C/rpc",
		"https://endpoints.omniatech.io/v1/avax/fuji/public",
		"https://api.avax-test.network/ext/bc/C/rpc",
	}

	BaseMainnetRPC = []string{
		"https://mainnet.base.org",
		"https://developer-access-mainnet.base.org",
		"https://base-rpc.publicnode.com",
	}

	BaseSepoliaRPC = []string{
		"https://base-sepolia-rpc.publicnode.com",
		"https://sepolia.base.org",
		"https://base-sepolia.gateway.tenderly.co",
	}

	XRPMainnetRPC = []string{
		"https://s1.ripple.com:51234",
		"https://xrplcluster.com",
	}

	XRPTestnetRPC = []string{
		"https://s.altnet.rippletest.net:51234",
	}
)

func GetAllRPCUrlByNetwork(chainId uint) []string {
	switch chainId {
	case ETH_MAINNET:
		return ETHMainnetRPC
	case ETH_SEPOLIA:
		return ETHSepoliaRPC
	case OP_MAINNET:
		return OPMainnetRPC
	case OP_SEPOLIA:
		return OPSepoliaRPC
	case BSC_MAINNET:
		return BSCMainnetRPC
	case BSC_TESTNET:
		return BSCTestnetRPC
	case ARBITRUM_ONE:
		return ArbitrumOneRPC
	case ARBITRUM_NOVA:
		return ArbitrumNovaRPC
	case ARBITRUM_SEPOLIA:
		return ArbitrumSepoliaRPC
	case POL_MAINNET:
		return PolMainnetRPC
	case POL_TESTNET:
		return PolTestnetRPC
	case AVAX_MAINNET:
		return AvaxMainnetRPC
	case AVAX_TESTNET:
		return AvaxTestnetRPC
	case BASE_MAINNET:
		return BaseMainnetRPC
	case BASE_SEPOLIA:
		return BaseSepoliaRPC
	case SOL_MAINNET:
		return SolanaMainnetRPC
	case SOL_DEVNET:
		return SolanaDevnetRpc
	case XRP_MAINNET:
		return XRPMainnetRPC
	case XRP_TESTNET:
		return XRPTestnetRPC
	default:
		return nil
	}
}

func GetGeneralRPCUrlByNetwork(chainId uint) string {
	switch chainId {
	case ETH_MAINNET:
		index := rand.IntN(len(ETHGeneralMainnetRPC))
		return ETHGeneralMainnetRPC[index]
	case ETH_SEPOLIA:
		index := rand.IntN(len(ETHGeneralSepoliaRPC))
		return ETHGeneralSepoliaRPC[index]
	default:
		return ""
	}

}

func GetAlchemyRPCUrlByNetwork(chainId uint) string {
	switch chainId {
	case ETH_MAINNET:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")))
		return "https://eth-mainnet.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")[index]
	case ETH_SEPOLIA:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")))
		return "https://eth-sepolia.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")[index]
	case BSC_MAINNET:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")))
		return "https://bnb-mainnet.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")[index]
	case BSC_TESTNET:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")))
		return "https://bnb-testnet.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")[index]
	case ARBITRUM_ONE:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")))
		return "https://arb-mainnet.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")[index]
	case ARBITRUM_NOVA:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")))
		return "https://arbnova-mainnet.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")[index]
	case ARBITRUM_SEPOLIA:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")))
		return "https://arb-sepolia.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")[index]
	case OP_MAINNET:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")))
		return "https://opt-mainnet.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")[index]
	case OP_SEPOLIA:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")))
		return "https://opt-sepolia.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")[index]
	case SOL_MAINNET:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")))
		return "https://solana-mainnet.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")[index]
	case SOL_DEVNET:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")))
		return "https://solana-devnet.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")[index]
	case POL_MAINNET:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")))
		return "https://polygon-mainnet.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")[index]
	case POL_TESTNET:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")))
		return "https://polygon-amoy.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")[index]
	case AVAX_MAINNET:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")))
		return "https://avax-mainnet.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")[index]
	case AVAX_TESTNET:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")))
		return "https://avax-fuji.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")[index]
	case BASE_MAINNET:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")))
		return "https://base-mainnet.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")[index]
	case BASE_SEPOLIA:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")))
		return "https://base-sepolia.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")[index]
	default:
		return ""
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

func GetRPCUrlByNetwork(chainId uint) string {
	switch chainId {
	case ETH_MAINNET:
		index := rand.IntN(len(ETHMainnetRPC))
		return ETHMainnetRPC[index]
	case ETH_SEPOLIA:
		index := rand.IntN(len(ETHSepoliaRPC))
		return ETHSepoliaRPC[index]
	case BSC_MAINNET:
		index := rand.IntN(len(BSCMainnetRPC))
		return BSCMainnetRPC[index]
	case BSC_TESTNET:
		index := rand.IntN(len(BSCTestnetRPC))
		return BSCTestnetRPC[index]
	case ARBITRUM_ONE:
		index := rand.IntN(len(ArbitrumOneRPC))
		return ArbitrumOneRPC[index]
	case ARBITRUM_NOVA:
		index := rand.IntN(len(ArbitrumNovaRPC))
		return ArbitrumNovaRPC[index]
	case ARBITRUM_SEPOLIA:
		index := rand.IntN(len(ArbitrumSepoliaRPC))
		return ArbitrumSepoliaRPC[index]
	case OP_MAINNET:
		index := rand.IntN(len(OPMainnetRPC))
		return OPMainnetRPC[index]
	case OP_SEPOLIA:
		index := rand.IntN(len(OPSepoliaRPC))
		return OPSepoliaRPC[index]
	case XRP_MAINNET:
		index := rand.IntN(len(XRPMainnetRPC))
		return XRPMainnetRPC[index]
	case XRP_TESTNET:
		index := rand.IntN(len(XRPTestnetRPC))
		return XRPTestnetRPC[index]
	case POL_MAINNET:
		index := rand.IntN(len(PolMainnetRPC))
		return PolMainnetRPC[index]
	case POL_TESTNET:
		index := rand.IntN(len(PolTestnetRPC))
		return PolTestnetRPC[index]
	case AVAX_MAINNET:
		index := rand.IntN(len(AvaxMainnetRPC))
		return AvaxMainnetRPC[index]
	case AVAX_TESTNET:
		index := rand.IntN(len(AvaxTestnetRPC))
		return AvaxTestnetRPC[index]
	case BASE_MAINNET:
		index := rand.IntN(len(BaseMainnetRPC))
		return BaseMainnetRPC[index]
	case BASE_SEPOLIA:
		index := rand.IntN(len(BaseSepoliaRPC))
		return BaseSepoliaRPC[index]
	case SOL_MAINNET:
		index := rand.IntN(len(SolanaMainnetRPC))
		return SolanaMainnetRPC[index]
	case SOL_DEVNET:
		index := rand.IntN(len(SolanaDevnetRpc))
		return SolanaDevnetRpc[index]
	default:
		return ""
	}
}

// get real inner tx(trace_debug) rpc url
func GetInnerTxRPCUrlByNetwork(chainId uint) string {
	switch chainId {
	case ETH_MAINNET:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxMainnetKey, ",")))
		return "https://eth-mainnet.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxMainnetKey, ",")[index]
	case ETH_SEPOLIA:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxTestnetKey, ",")))
		return "https://eth-sepolia.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxTestnetKey, ",")[index]
	case BSC_MAINNET:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxMainnetKey, ",")))
		return "https://bnb-mainnet.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxMainnetKey, ",")[index]
	case BSC_TESTNET:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxTestnetKey, ",")))
		return "https://bnb-testnet.g.alchemy.com/v2/" + strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxTestnetKey, ",")[index]
	default:
		return ""
	}
}

func GetRandomAlchemyKey(isMainnet bool) string {
	if isMainnet {
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")[index]
	} else {
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")[index]
	}
}

func GetRandomInnertxAlchemyKey(isMainnet bool) string {
	if isMainnet {
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxMainnetKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxMainnetKey, ",")[index]
	} else {
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxTestnetKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxTestnetKey, ",")[index]
	}
}
