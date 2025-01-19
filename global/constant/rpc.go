package constant

import (
	"math/rand"
	"node/global"
	"node/model/node/request"
	"node/model/node/response"
	"node/utils"
	"strings"
	"time"
)

var (
	ETHAlchemyMainnetRPC = []string{
		"https://eth-mainnet.g.alchemy.com/v2/" + GetRandomAlchemyKey(true),
	}

	ETHAlchemyGoerliRPC = []string{
		"https://eth-goerli.g.alchemy.com/v2/" + GetRandomAlchemyKey(false),
	}

	ETHAlchemySepoliaRPC = []string{
		"https://eth-sepolia.g.alchemy.com/v2/" + GetRandomAlchemyKey(false),
	}

	ETHMainnetRPC = []string{
		"https://ethereum-rpc.publicnode.com",
		"https://eth-mainnet.g.alchemy.com/v2/" + GetRandomAlchemyKey(true),
	}

	ETHInnerTxMainnetRPC = []string{
		// "https://eth.llamarpc.com",
		// "https://eth-pokt.nodies.app",
		// "https://eth.merkle.io",
		// "https://eth.nodeconnect.org",
		// "https://gateway.subquery.network/rpc/eth",
		// "https://ethereum.rpc.subquery.network/public",
		"https://eth-mainnet.g.alchemy.com/v2/" + GetRandomInnertxAlchemyKey(true),
	}

	ETHGeneralMainnetRPC = []string{
		"https://ethereum-rpc.publicnode.com",
	}

	ETHGoerliRPC = []string{
		"https://rpc.ankr.com/eth_goerli",
		"https://rpc.tornadoeth.cash/goerli",
	}

	ETHGeneralGoerliRPC = []string{
		"https://rpc.ankr.com/eth_goerli",
		"https://rpc.tornadoeth.cash/goerli",
	}

	ETHSepoliaRPC = []string{
		// "https://eth-sepolia.g.alchemy.com/v2/" + GetRandomAlchemyKey(false),
		"https://ethereum-sepolia.publicnode.com",
	}

	ETHInnerTxSepoliaRPC = []string{
		"https://eth-sepolia.g.alchemy.com/v2/" + GetRandomInnertxAlchemyKey(false),
	}

	ETHGeneralSepoliaRPC = []string{
		"https://ethereum-sepolia.publicnode.com",
	}

	BSCTestnetRPC = []string{
		"https://data-seed-prebsc-1-s1.binance.org:8545",
		"https://data-seed-prebsc-2-s1.binance.org:8545",
		"https://data-seed-prebsc-1-s2.binance.org:8545",
		"https://data-seed-prebsc-2-s2.binance.org:8545",
		"https://data-seed-prebsc-1-s3.binance.org:8545",
		"https://data-seed-prebsc-2-s3.binance.org:8545",
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

	OPMainnetRPC = []string{
		// "https://opt-mainnet.g.alchemy.com/v2/" + GetRandomAlchemyKey(true),
		// "https://mainnet.optimism.io",
		"https://optimism-rpc.publicnode.com",
		"https://op-pokt.nodies.app",
		"https://1rpc.io/op",
	}

	OPGoerliRPC = []string{
		// "https://opt-goerli.g.alchemy.com/v2/" + GetRandomAlchemyKey(false),
		"https://goerli.optimism.io",
	}

	OPSepoliaRPC = []string{
		"https://opt-sepolia.g.alchemy.com/v2/" + GetRandomAlchemyKey(false),
		"https://sepolia.optimism.io",
		"https://optimism-sepolia.drpc.org",
		"https://endpoints.omniatech.io/v1/op/sepolia/public",
	}

	ArbitrumOneRPC = []string{
		// "https://arb-mainnet.g.alchemy.com/v2/" + GetRandomAlchemyKey(true),
		"https://arb1.arbitrum.io/rpc",
		// "https://arbitrum.llamarpc.com",
		"https://arbitrum.meowrpc.com",
		// "https://arbitrum.drpc.org",
		// "https://arbitrum-one.publicnode.com",
		// "https://arbitrum-one-rpc.publicnode.com",
		// "https://1rpc.io/arb",
	}

	ArbitrumNovaRPC = []string{
		"https://nova.arbitrum.io/rpc",
	}

	ArbitrumGoerliRPC = []string{
		// "https://arb-goerli.g.alchemy.com/v2/" + GetRandomAlchemyKey(false),
		"https://goerli-rollup.arbitrum.io/rpc",
	}

	ArbitrumSepoliaRPC = []string{
		"https://sepolia-rollup.arbitrum.io/rpc",
		"https://arbitrum-sepolia.blockpi.network/v1/rpc/public",
	}
)

func GetAllRPCUrlByNetwork(chainId uint) []string {
	switch chainId {
	case ETH_MAINNET:
		return ETHMainnetRPC
	case ETH_GOERLI:
		return ETHGoerliRPC
	case ETH_SEPOLIA:
		return ETHSepoliaRPC
	case OP_MAINNET:
		return OPMainnetRPC
	case OP_GOERLI:
		return OPGoerliRPC
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
	case ARBITRUM_GOERLI:
		return ArbitrumGoerliRPC
	case ARBITRUM_SEPOLIA:
		return ArbitrumSepoliaRPC
	}

	return nil
}

func GetGeneralRPCUrlByNetwork(chainId uint) string {
	rand.Seed(time.Now().UnixNano())

	switch chainId {
	case ETH_MAINNET:
		index := rand.Intn(len(ETHGeneralMainnetRPC))
		return ETHGeneralMainnetRPC[index]
	case ETH_GOERLI:
		index := rand.Intn(len(ETHGeneralGoerliRPC))
		return ETHGeneralGoerliRPC[index]
	case ETH_SEPOLIA:
		index := rand.Intn(len(ETHGeneralSepoliaRPC))
		return ETHGeneralSepoliaRPC[index]
	}

	return ""
}

func GetAlchemyRPCUrlByNetwork(chainId uint) string {
	rand.Seed(time.Now().UnixNano())

	switch chainId {
	case ETH_MAINNET:
		index := rand.Intn(len(ETHAlchemyMainnetRPC))
		return ETHAlchemyMainnetRPC[index]
	case ETH_GOERLI:
		index := rand.Intn(len(ETHAlchemyGoerliRPC))
		return ETHAlchemyGoerliRPC[index]
	case ETH_SEPOLIA:
		index := rand.Intn(len(ETHAlchemySepoliaRPC))
		return ETHAlchemySepoliaRPC[index]
	}

	return ""
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

// get real rpc url
func GetRPCUrlByNetwork(chainId uint) string {
	switch chainId {
	case ETH_MAINNET:
		return GetRealRpcByArray(ETHMainnetRPC)
	case ETH_GOERLI:
		return GetRealRpcByArray(ETHGoerliRPC)
	case ETH_SEPOLIA:
		return GetRealRpcByArray(ETHSepoliaRPC)
	case OP_MAINNET:
		return GetRealRpcByArray(OPMainnetRPC)
	case OP_GOERLI:
		return GetRealRpcByArray(OPGoerliRPC)
	case OP_SEPOLIA:
		return GetRealRpcByArray(OPSepoliaRPC)
	case BSC_MAINNET:
		return GetRealRpcByArray(BSCMainnetRPC)
	case BSC_TESTNET:
		return GetRealRpcByArray(BSCTestnetRPC)
	case ARBITRUM_ONE:
		rand.Seed(time.Now().UnixNano())
		index := rand.Intn(len(ArbitrumOneRPC))
		return ArbitrumOneRPC[index]
	case ARBITRUM_NOVA:
		return GetRealRpcByArray(ArbitrumNovaRPC)
	case ARBITRUM_GOERLI:
		return GetRealRpcByArray(ArbitrumGoerliRPC)
	case ARBITRUM_SEPOLIA:
		return GetRealRpcByArray(ArbitrumSepoliaRPC)
	}

	return ""
}

// get real inner tx(trace_debug) rpc url
func GetInnerTxRPCUrlByNetwork(chainId uint) string {
	rand.Seed(time.Now().UnixNano())

	switch chainId {
	case ETH_MAINNET:
		index := rand.Intn(len(ETHInnerTxMainnetRPC))
		return ETHInnerTxMainnetRPC[index]
		// return GetRealRpcByArray(ETHInnerTxMainnetRPC)
	case ETH_SEPOLIA:
		index := rand.Intn(len(ETHInnerTxSepoliaRPC))
		return ETHInnerTxSepoliaRPC[index]
		// return GetRealRpcByArray(ETHInnerTxSepoliaRPC)
	}

	return ""
}

func GetRandomAlchemyKey(isMainnet bool) string {
	rand.Seed(time.Now().UnixNano())

	if isMainnet {
		index := rand.Intn(len(strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.AlchemyMainnetKey, ",")[index]
	} else {
		index := rand.Intn(len(strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.AlchemyTestnetKey, ",")[index]
	}
}

func GetRandomInnertxAlchemyKey(isMainnet bool) string {
	rand.Seed(time.Now().UnixNano())

	if isMainnet {
		index := rand.Intn(len(strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxMainnetKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxMainnetKey, ",")[index]
	} else {
		index := rand.Intn(len(strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxTestnetKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.AlchemyInnerTxTestnetKey, ",")[index]
	}
}
