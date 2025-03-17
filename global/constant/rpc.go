package constant

import (
	"math/rand/v2"
	"node/global"
	"node/model/node/request"
	"node/model/node/response"
	"node/utils"
	"strings"
)

var (
	// ETHAlchemyMainnetRPC = []string{
	// 	"https://eth-mainnet.g.alchemy.com/v2/" + GetRandomAlchemyKey(true),
	// }

	// ETHAlchemySepoliaRPC = []string{
	// 	"https://eth-sepolia.g.alchemy.com/v2/" + GetRandomAlchemyKey(false),
	// }

	// ETHInnerTxMainnetRPC = []string{
	// 	// "https://eth.llamarpc.com",
	// 	// "https://eth-pokt.nodies.app",
	// 	// "https://eth.merkle.io",
	// 	// "https://eth.nodeconnect.org",
	// 	// "https://gateway.subquery.network/rpc/eth",
	// 	// "https://ethereum.rpc.subquery.network/public",
	// 	"https://eth-mainnet.g.alchemy.com/v2/" + GetRandomInnertxAlchemyKey(true),
	// }

	// ETHInnerTxSepoliaRPC = []string{
	// 	"https://eth-sepolia.g.alchemy.com/v2/" + GetRandomInnertxAlchemyKey(false),
	// }

	ETHGeneralMainnetRPC = []string{
		"https://ethereum-rpc.publicnode.com",
	}

	ETHGeneralSepoliaRPC = []string{
		"https://ethereum-sepolia.publicnode.com",
	}

	ETHMainnetRPC = []string{
		// 	"https://eth-mainnet.g.alchemy.com/v2/" + GetRandomAlchemyKey(true),
		"https://ethereum-rpc.publicnode.com",
	}

	ETHSepoliaRPC = []string{
		// "https://eth-sepolia.g.alchemy.com/v2/" + GetRandomAlchemyKey(false),
		"https://ethereum-sepolia.publicnode.com",
	}

	BSCMainnetRPC = []string{
		"https://bsc-dataseed1.binance.org",
		"https://bsc-dataseed2.binance.org",
		"https://bsc-dataseed3.binance.org",
		"https://bsc-dataseed4.binance.org",
		"https://bsc-dataseed1.defibit.io",
		"https://bsc-dataseed2.defibit.io",
		"https://bsc-dataseed3.defibit.io",
		"https://bsc-dataseed4.defibit.io",
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
		// "https://opt-mainnet.g.alchemy.com/v2/" + GetRandomAlchemyKey(true),
		"https://mainnet.optimism.io",
		"https://optimism-rpc.publicnode.com",
		"https://op-pokt.nodies.app",
		"https://1rpc.io/op",
	}

	OPSepoliaRPC = []string{
		// "https://opt-sepolia.g.alchemy.com/v2/" + GetRandomAlchemyKey(false),
		"https://sepolia.optimism.io",
		"https://optimism-sepolia.drpc.org",
		"https://endpoints.omniatech.io/v1/op/sepolia/public",
	}

	ArbitrumOneRPC = []string{
		// "https://arb-mainnet.g.alchemy.com/v2/" + GetRandomAlchemyKey(true),
		"https://arb1.arbitrum.io/rpc",
		"https://arbitrum.llamarpc.com",
		"https://arbitrum.meowrpc.com",
		"https://arbitrum.drpc.org",
		"https://arbitrum-one.publicnode.com",
		"https://arbitrum-one-rpc.publicnode.com",
		"https://1rpc.io/arb",
	}

	ArbitrumNovaRPC = []string{
		"https://nova.arbitrum.io/rpc",
	}

	ArbitrumSepoliaRPC = []string{
		"https://sepolia-rollup.arbitrum.io/rpc",
	}

	SolanaMainnetRPC = []string{
		"https://api.mainnet-beta.solana.com",
		"https://rpc.ankr.com/solana",
	}

	SolanaDevnetRpc = []string{
		"https://api.devnet.solana.com",
		// "https://rpc.ankr.com/solana_devnet",
	}

	PolMainnetRPC = []string{
		"https://polygon-bor-rpc.publicnode.com",
		"https://polygon-pokt.nodies.app",
		"https://1rpc.io/matic",
	}

	PolTestnetRPC = []string{
		"https://polygon-amoy-bor-rpc.publicnode.com",
		"https://rpc-amoy.polygon.technology",
		"https://polygon-amoy.drpc.org",
	}

	AvaxMainnetRPC = []string{
		"https://avalanche-c-chain-rpc.publicnode.com",
		"https://avalanche.drpc.org",
		"https://avax.meowrpc.com",
		"https://1rpc.io/avax/c",
	}

	AvaxTestnetRPC = []string{
		"https://ava-testnet.public.blastapi.io/ext/bc/C/rpc",
		"https://endpoints.omniatech.io/v1/avax/fuji/public",
		"https://api.avax-test.network/ext/bc/C/rpc",
		"https://rpc.ankr.com/avalanche_fuji",
	}

	BaseMainnetRPC = []string{
		"https://base.llamarpc.com",
		"https://mainnet.base.org",
		"https://developer-access-mainnet.base.org",
		"https://1rpc.io/base",
		"https://base-pokt.nodies.app",
		"https://base.meowrpc.com",
		"https://base-rpc.publicnode.com",
		"https://base.drpc.org",
	}

	BaseSepoliaRPC = []string{
		"https://base-sepolia-rpc.publicnode.com",
		"https://sepolia.base.org",
		"https://base-sepolia.gateway.tenderly.co",
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
	default:
		return ""
	}

}

func GetRealRpcByArray(rpcs []string) string {
	for _, rpc := range rpcs {
		client.URL = rpc
		var rpcBlockInfo response.RPCBlockInfo
		var jsonRpcRequest request.JsonRpcRequest
		jsonRpcRequest.Id = 1
		jsonRpcRequest.Jsonrpc = "2.0"
		jsonRpcRequest.Method = "eth_getBlockByNumber"
		jsonRpcRequest.Params = []interface{}{"latest", false}
		err := client.HTTPPost(jsonRpcRequest, &rpcBlockInfo)
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
		return GetRealRpcByArray(ETHMainnetRPC)
	case ETH_SEPOLIA:
		return GetRealRpcByArray(ETHSepoliaRPC)
	case BSC_MAINNET:
		return GetRealRpcByArray(BSCMainnetRPC)
	case BSC_TESTNET:
		return GetRealRpcByArray(BSCTestnetRPC)
	case OP_MAINNET:
		return GetRealRpcByArray(OPMainnetRPC)
	case OP_SEPOLIA:
		return GetRealRpcByArray(OPSepoliaRPC)
	case ARBITRUM_ONE:
		return GetRealRpcByArray(ArbitrumOneRPC)
	case ARBITRUM_NOVA:
		return GetRealRpcByArray(ArbitrumNovaRPC)
	case ARBITRUM_SEPOLIA:
		return GetRealRpcByArray(ArbitrumSepoliaRPC)
	case SOL_MAINNET:
		index := rand.IntN(len(SolanaMainnetRPC))
		return SolanaMainnetRPC[index]
	case SOL_DEVNET:
		index := rand.IntN(len(SolanaDevnetRpc))
		return SolanaDevnetRpc[index]
	case POL_MAINNET:
		return GetRealRpcByArray(PolMainnetRPC)
	case POL_TESTNET:
		return GetRealRpcByArray(PolTestnetRPC)
	case AVAX_MAINNET:
		return GetRealRpcByArray(AvaxMainnetRPC)
	case AVAX_TESTNET:
		return GetRealRpcByArray(AvaxTestnetRPC)
	case BASE_MAINNET:
		return GetRealRpcByArray(BaseMainnetRPC)
	case BASE_SEPOLIA:
		return GetRealRpcByArray(BaseSepoliaRPC)
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

func GetRandomInnertxAlchemyKey(isMainnet bool) string {
	if isMainnet {
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxMainnetKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxMainnetKey, ",")[index]
	} else {
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxTestnetKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxTestnetKey, ",")[index]
	}
}
