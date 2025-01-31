package constant

import (
	"context"
	"node/global"
	"node/model/node/request"
	"node/model/node/response"
	NODE_Client "node/utils/http"
	"strings"

	solanaCommon "github.com/blocto/solana-go-sdk/common"
	solanaRpc "github.com/blocto/solana-go-sdk/rpc"
	"github.com/btcsuite/btcd/btcutil"
	btcCfg "github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/common"
	ltcCfg "github.com/ltcsuite/ltcd/chaincfg"
	"github.com/ltcsuite/ltcd/ltcutil"
	tonAddress "github.com/xssnick/tonutils-go/address"
)

var ChainId = map[uint]string{
	1:          "Ethereum Mainnet",
	2:          "Bitcoin Mainnet",
	3:          "Bitcoin Testnet",
	5:          "Ethereum Goerli",
	6:          "Litecoin Mainnet",
	7:          "Litecoin Testnet",
	11155111:   "Ethereum Sepolia",
	101:        "Solana Mainnet",
	103:        "Solana Devnet",
	56:         "Binance Smart Chain Mainnet",
	97:         "Binance Smart Chain Testnet",
	1100:       "Ton Mainnet",
	1101:       "Ton Testnet",
	204:        "OpBNB Mainnet",
	5611:       "opBNB Testnet",
	137:        "Polygon Mainnet",
	80001:      "Polygon Testnet(Mumbai)",
	43114:      "Avalanche Mainnet(C-Chain)",
	43113:      "Avalanche Testnet (FUji)",
	250:        "Fantom Mainnet(Opera)",
	4002:       "Fantom Testnet",
	42161:      "Arbitrum One",
	42170:      "Arbitrum Nova",
	421613:     "Arbitrum Goerli",
	421614:     "Arbitrum Sepolia",
	1284:       "Moonbeam",
	1285:       "Moonriver",
	1287:       "Moonbase Alpha",
	1666600000: "Harmony Mainnet",
	1666700000: "Harmony Testnet",
	128:        "Huobi ECO Chain Mainnet",
	256:        "Huobi ECO Chain Testnet",
	1313161554: "Aurora Mainnet",
	1313161555: "Aurora Testnet",
	10:         "Optimism Mainnet",
	420:        "Optimism Goerli",
	11155420:   "Optimism Sepolia",
	321:        "KCC Mainnet",
	322:        "KCC Testnet",
	210425:     "PlatON Mainnet",
	2206132:    "PlatON Testnet",
	728126428:  "Tron Mainnet",
	2494104990: "Tron Shasta",
	3448148188: "Tron Nile",
	66:         "OKTC Mainnet",
	65:         "OKTC Testnet",
	195:        "OKBC Testnet",
	108:        "ThunderCore Mainnet",
	18:         "ThunderCore Testnet",
	25:         "Cronos Mainnet",
	338:        "Cronos Testnet",
	42262:      "OasisEmerald Mainnet",
	42261:      "OasisEmerald Testnet",
	100:        "Gnosis Mainnet",
	10200:      "Gnosis Testnet",
	42220:      "Celo Mainnet",
	44787:      "Celo Testnet",
	8217:       "Klaytn Mainnet",
	1001:       "Klavtn Testnet",
	324:        "zkSync Era Mainnet",
	280:        "zkSvnc Era Testnet",
	1088:       "Metis Mainnet",
	599:        "Metis Testnet",
	534351:     "Scroll Sepolia Testnet",
	534353:     "Scroll Alpha Testnet",
	1030:       "Conflux eSpace Mainnet",
	71:         "Conflux eSpace Testnet",
	22776:      "MAP Protocol Mainnet",
	212:        "MAP Protocol Testnet",
	8453:       "Base Mainnet",
	84531:      "Base Goerli Testnet",
	84532:      "Base Sepolia Testnet",
	59144:      "Linea Mainnet",
	59140:      "Linea Goerli Testet",
	5000:       "Mantle Mainnet",
	5001:       "Mantle Testnet",
	91715:      "Combo Testnet",
	12009:      "zkMeta Testnet",
	167005:     "Taiko Testnet",
	7777777:    "Zora Mainnet",
	999:        "Zora Mainnet",
	424:        "PGN Mainnet",
	58008:      "PGN Testnet",
	1482601649: "SKALE Nebula Mainnet",
	3441005:    "Manta Testnet",
	12015:      "ReadON Testnet",
	12021:      "GasZero Goerli",
}

var (
	ETH_MAINNET      uint = 1
	BTC_MAINNET      uint = 2
	BTC_TESTNET      uint = 3
	ETH_GOERLI       uint = 5
	LTC_MAINNET      uint = 6
	LTC_TESTNET      uint = 7
	OP_MAINNET       uint = 10
	BSC_MAINNET      uint = 56
	BSC_TESTNET      uint = 97
	OP_GOERLI        uint = 420
	ARBITRUM_ONE     uint = 42161
	ARBITRUM_NOVA    uint = 42170
	ARBITRUM_GOERLI  uint = 421613
	ARBITRUM_SEPOLIA uint = 421614
	ETH_SEPOLIA      uint = 11155111
	OP_SEPOLIA       uint = 11155420
	TRON_MAINNET     uint = 728126428
	TRON_NILE        uint = 3448148188
	SOL_MAINNET      uint = 101
	SOL_DEVNET       uint = 103
	TON_MAINNET      uint = 1100
	TON_TESTNET      uint = 1101

	JoinSweep = []uint{
		ETH_MAINNET,
		ETH_GOERLI,
		ETH_SEPOLIA,
		BTC_MAINNET,
		BTC_TESTNET,
		BSC_MAINNET,
		BSC_TESTNET,
		ARBITRUM_ONE,
		ARBITRUM_NOVA,
		ARBITRUM_GOERLI,
		ARBITRUM_SEPOLIA,
		TRON_MAINNET,
		TRON_NILE,
		LTC_MAINNET,
		LTC_TESTNET,
		OP_MAINNET,
		OP_SEPOLIA,
		SOL_MAINNET,
		SOL_DEVNET,
		TON_MAINNET,
		TON_TESTNET,
	}

	LikeMainnetEthChain = []uint{
		ETH_MAINNET,
		BSC_MAINNET,
		ARBITRUM_ONE,
		ARBITRUM_NOVA,
		OP_MAINNET,
	}

	LikeTestnetEthChain = []uint{
		ETH_GOERLI,
		ETH_SEPOLIA,
		BSC_TESTNET,
		ARBITRUM_GOERLI,
		ARBITRUM_SEPOLIA,
		OP_SEPOLIA,
	}

	LikeEthChain = []uint{
		ETH_MAINNET,
		BSC_MAINNET,
		ARBITRUM_ONE,
		ARBITRUM_NOVA,
		OP_MAINNET,
		ETH_GOERLI,
		ETH_SEPOLIA,
		BSC_TESTNET,
		ARBITRUM_GOERLI,
		ARBITRUM_SEPOLIA,
		OP_SEPOLIA,
	}

	LikeMainnetTronChain = []uint{
		TRON_MAINNET,
	}

	LikeTestnetTronChain = []uint{
		TRON_NILE,
	}

	LikeTronChain = []uint{
		TRON_MAINNET,
		TRON_NILE,
	}

	LikeMainnetBtcChain = []uint{
		BTC_MAINNET,
	}

	LikeTestnetBtcChain = []uint{
		BTC_TESTNET,
	}

	LikeBtcChain = []uint{
		BTC_MAINNET,
		BTC_TESTNET,
	}

	LikeMainnetLtcChain = []uint{
		LTC_MAINNET,
	}

	LikeTestnetLtcChain = []uint{
		LTC_TESTNET,
	}

	LikeLtcChain = []uint{
		LTC_MAINNET,
		LTC_TESTNET,
	}

	MainnetChain = []uint{
		BTC_MAINNET,
		LTC_MAINNET,
		ETH_MAINNET,
		OP_MAINNET,
		BSC_MAINNET,
		ARBITRUM_ONE,
		ARBITRUM_NOVA,
		TRON_MAINNET,
		SOL_MAINNET,
		TON_MAINNET,
	}

	TestnetChain = []uint{
		BTC_TESTNET,
		LTC_TESTNET,
		ETH_SEPOLIA,
		OP_SEPOLIA,
		BSC_TESTNET,
		ARBITRUM_SEPOLIA,
		TRON_NILE,
		SOL_DEVNET,
		TON_TESTNET,
	}
)

func IsNetworkSupport(chainId uint) bool {
	return ChainId[chainId] != ""
}

func IsMainnetSupport(chainId uint) bool {
	for _, v := range MainnetChain {
		if chainId == v {
			return true
		}
	}

	return false
}

func IsTestnetSupport(chainId uint) bool {
	for _, v := range TestnetChain {
		if chainId == v {
			return true
		}
	}

	return false
}

func IsNetworkLikeEth(chainId uint) bool {
	for _, v := range LikeEthChain {
		if chainId == v {
			return true
		}
	}

	return false
}

func IsNetworkLikeTron(chainId uint) bool {
	for _, v := range LikeTronChain {
		if chainId == v {
			return true
		}
	}

	return false
}

func IsNetworkLikeBtc(chainId uint) bool {
	for _, v := range LikeBtcChain {
		if chainId == v {
			return true
		}
	}

	return false
}

func IsNetworkLikeLtc(chainId uint) bool {
	for _, v := range LikeLtcChain {
		if chainId == v {
			return true
		}
	}

	return false
}

func IsAddressSupport(chainId uint, address string) bool {
	if !IsNetworkSupport(chainId) {
		return false
	}

	switch chainId {
	case ETH_MAINNET,
		ETH_GOERLI,
		ETH_SEPOLIA,
		BSC_MAINNET,
		BSC_TESTNET,
		ARBITRUM_ONE,
		ARBITRUM_NOVA,
		ARBITRUM_GOERLI,
		ARBITRUM_SEPOLIA,
		OP_MAINNET,
		OP_SEPOLIA:
		return common.IsHexAddress(address)
	case BTC_MAINNET:
		_, err := btcutil.DecodeAddress(address, &btcCfg.MainNetParams)
		if err != nil {
			return false
		}
		return true
	case BTC_TESTNET:
		_, err := btcutil.DecodeAddress(address, &btcCfg.TestNet3Params)
		if err != nil {
			return false
		}
		return true
	case LTC_MAINNET:
		_, err := ltcutil.DecodeAddress(address, &ltcCfg.MainNetParams)
		if err != nil {
			return false
		}
		return true
	case LTC_TESTNET:
		_, err := ltcutil.DecodeAddress(address, &ltcCfg.TestNet4Params)
		if err != nil {
			return false
		}
		return true
	case TRON_MAINNET, TRON_NILE:
		resultVal, _ := TronValidateAddress(chainId, address)
		return resultVal
	case SOL_MAINNET:
		resultVal, t := SolanaValidateAddress(solanaRpc.MainnetRPCEndpoint, address)
		if resultVal && t == "address" {
			return true
		}
		return false
	case SOL_DEVNET:
		resultVal, t := SolanaValidateAddress(solanaRpc.DevnetRPCEndpoint, address)
		if resultVal && t == "address" {
			return true
		}
		return false
	case TON_MAINNET, TON_TESTNET:
		resultVal, err := tonAddress.ParseAddr(address)
		if err != nil {
			return false
		}
		if resultVal.Type() == tonAddress.StdAddress {
			return true
		}
		return false
	}

	return false
}

var (
	client NODE_Client.Client
)

func IsAddressContractSupport(chainId uint, address string) bool {
	if !IsNetworkSupport(chainId) {
		return false
	}

	switch chainId {
	case ETH_MAINNET,
		ETH_GOERLI,
		ETH_SEPOLIA,
		BSC_MAINNET,
		BSC_TESTNET,
		ARBITRUM_ONE,
		ARBITRUM_NOVA,
		ARBITRUM_GOERLI,
		ARBITRUM_SEPOLIA,
		OP_MAINNET,
		OP_SEPOLIA:
		client.URL = GetRPCUrlByNetwork(chainId)
		var rpcGeneral response.RPCGeneral
		var jsonRpcRequest request.JsonRpcRequest
		jsonRpcRequest.Id = 1
		jsonRpcRequest.Jsonrpc = "2.0"
		jsonRpcRequest.Method = "eth_getCode"
		jsonRpcRequest.Params = []interface{}{address, "latest"}

		err := client.HTTPPost(jsonRpcRequest, &rpcGeneral)
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			return false
		}

		if rpcGeneral.Result != "0x" {
			return true
		}
	case TRON_MAINNET, TRON_NILE:
		resultVal, _ := TronValidateContratAddress(chainId, address)
		return resultVal
	case SOL_MAINNET:
		resultVal, t := SolanaValidateAddress(solanaRpc.MainnetRPCEndpoint, address)
		if resultVal && t == "contract" {
			return true
		}
		return false
	case SOL_DEVNET:
		resultVal, t := SolanaValidateAddress(solanaRpc.DevnetRPCEndpoint, address)
		if resultVal && t == "contract" {
			return true
		}
		return false
	case TON_MAINNET, TON_TESTNET:
		resultVal, err := tonAddress.ParseAddr(address)
		if err != nil {
			return false
		}
		if resultVal.Type() == tonAddress.VarAddress {
			return true
		}
		return false
	}

	return false
}

func GetChainName(chainId uint) string {
	return ChainId[chainId]
}

func AddressToLower(chainId uint, address string) string {
	if !IsNetworkSupport(chainId) {
		return ""
	}

	switch chainId {
	case ETH_MAINNET,
		ETH_GOERLI,
		ETH_SEPOLIA,
		BSC_MAINNET,
		BSC_TESTNET,
		ARBITRUM_ONE,
		ARBITRUM_NOVA,
		ARBITRUM_GOERLI,
		ARBITRUM_SEPOLIA,
		OP_MAINNET,
		OP_SEPOLIA:
		return strings.ToLower(address)
	}

	return address
}

func TronValidateAddress(chainId uint, address string) (bool, string) {
	client.URL = TronValidateAddressByNetwork(chainId)
	client.Headers = map[string]string{
		"TRON-PRO-API-KEY": GetRandomHTTPKeyByNetwork(chainId),
	}

	var addressRequest request.TronValidateAddressRequest
	addressRequest.Address = address
	addressRequest.Visible = true
	var addressResponse response.TronValidateAddressResponse
	err := client.HTTPPost(addressRequest, &addressResponse)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return false, ""
	}

	return addressResponse.Result, addressResponse.Message
}

func TronValidateContratAddress(chainId uint, address string) (bool, string) {
	client.URL = TronValidateContractAddressByNetwork(chainId)
	client.Headers = map[string]string{
		"TRON-PRO-API-KEY": GetRandomHTTPKeyByNetwork(chainId),
	}

	var contractRequest request.TronContractRequest
	contractRequest.Value = address
	contractRequest.Visible = true
	var contractResponse response.TronContractResponse
	err := client.HTTPPost(contractRequest, &contractResponse)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return false, ""
	}

	if contractResponse.Bytecode != "" && contractResponse.ContractAddress != "" {
		return true, contractResponse.ContractAddress
	}

	return false, ""
}

func SolanaValidateAddress(rpc, address string) (bool, string) {
	client := solanaRpc.NewRpcClient(rpc)

	pubKey := solanaCommon.PublicKeyFromString(address)

	accountInfo, err := client.GetAccountInfo(context.Background(), pubKey.ToBase58())
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return false, ""
	}

	if accountInfo.GetResult().Value.Owner == "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
		return true, "contract"
	} else {
		return true, "address"
	}
}
