package constant

import (
	"node/global"
	"node/model/node/request"
	"node/model/node/response"
	NODE_Client "node/utils/http"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	btcCfg "github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gagliardetto/solana-go"
	bchCfg "github.com/gcash/bchd/chaincfg"
	"github.com/gcash/bchutil"

	// ltcCfg "github.com/ltcsuite/ltcd/chaincfg"
	// "github.com/ltcsuite/ltcd/ltcutil"
	"github.com/xrpscan/xrpl-go"
	tonAddress "github.com/xssnick/tonutils-go/address"
)

var ChainId = map[uint]string{
	1:          "Ethereum Mainnet",
	2:          "Bitcoin Mainnet",
	3:          "Bitcoin Testnet",
	5:          "Ethereum Goerli",
	6:          "Litecoin Mainnet",
	7:          "Litecoin Testnet",
	8:          "Xrp Mainnet",
	9:          "Xrp Testnet",
	11:         "Bitcoin Cash Mainnet",
	12:         "Bitcoin Cash Testnet",
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
	80002:      "Polygon Testnet(Amoy)",
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
	ETH_MAINNET uint = 1
	ETH_SEPOLIA uint = 11155111

	BTC_MAINNET uint = 2
	BTC_TESTNET uint = 3

	LTC_MAINNET uint = 6
	LTC_TESTNET uint = 7

	BSC_MAINNET uint = 56
	BSC_TESTNET uint = 97

	ARBITRUM_ONE     uint = 42161
	ARBITRUM_NOVA    uint = 42170
	ARBITRUM_SEPOLIA uint = 421614

	OP_MAINNET uint = 10
	OP_SEPOLIA uint = 11155420

	TRON_MAINNET uint = 728126428
	TRON_NILE    uint = 3448148188

	SOL_MAINNET uint = 101
	SOL_DEVNET  uint = 103

	TON_MAINNET uint = 1100
	TON_TESTNET uint = 1101

	XRP_MAINNET uint = 8
	XRP_TESTNET uint = 9

	BCH_MAINNET uint = 11
	BCH_TESTNET uint = 12

	POL_MAINNET uint = 137
	POL_TESTNET uint = 80002

	AVAX_MAINNET uint = 43114
	AVAX_TESTNET uint = 43113

	BASE_MAINNET uint = 8453
	BASE_SEPOLIA uint = 84532

	JoinSweep = []uint{
		ETH_MAINNET,
		ETH_SEPOLIA,
		BTC_MAINNET,
		BTC_TESTNET,
		BSC_MAINNET,
		BSC_TESTNET,
		ARBITRUM_ONE,
		ARBITRUM_NOVA,
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
		XRP_MAINNET,
		XRP_TESTNET,
		BCH_MAINNET,
		BCH_TESTNET,
		POL_MAINNET,
		POL_TESTNET,
		AVAX_MAINNET,
		AVAX_TESTNET,
		BASE_MAINNET,
		BASE_SEPOLIA,
	}

	MainnetEthChain = []uint{
		ETH_MAINNET,
		BSC_MAINNET,
		ARBITRUM_ONE,
		ARBITRUM_NOVA,
		OP_MAINNET,
		POL_MAINNET,
		AVAX_MAINNET,
		BASE_MAINNET,
	}

	TestnetEthChain = []uint{
		ETH_SEPOLIA,
		BSC_TESTNET,
		ARBITRUM_SEPOLIA,
		OP_SEPOLIA,
		POL_TESTNET,
		AVAX_TESTNET,
		BASE_SEPOLIA,
	}

	MainnetTronChain = []uint{
		TRON_MAINNET,
	}

	TestnetTronChain = []uint{
		TRON_NILE,
	}

	MainnetBtcChain = []uint{
		BTC_MAINNET,
	}

	TestnetBtcChain = []uint{
		BTC_TESTNET,
	}

	MainnetLtcChain = []uint{
		LTC_MAINNET,
	}

	TestnetLtcChain = []uint{
		LTC_TESTNET,
	}

	MainnetSolChain = []uint{
		SOL_MAINNET,
	}

	TestnetSolChain = []uint{
		SOL_DEVNET,
	}

	MainnetTonChain = []uint{
		TON_MAINNET,
	}

	TestnetTonChain = []uint{
		TON_TESTNET,
	}

	MainnetXrpChain = []uint{
		XRP_MAINNET,
	}

	TestnetXrpChain = []uint{
		XRP_TESTNET,
	}

	MainnetBchChain = []uint{
		BCH_MAINNET,
	}

	TestnetBchChain = []uint{
		BCH_TESTNET,
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
		XRP_MAINNET,
		BCH_MAINNET,
		POL_MAINNET,
		AVAX_MAINNET,
		BASE_MAINNET,
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
		XRP_TESTNET,
		BCH_TESTNET,
		POL_TESTNET,
		AVAX_TESTNET,
		BASE_SEPOLIA,
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

func IsAddressSupport(chainId uint, address string) bool {
	if !IsNetworkSupport(chainId) {
		return false
	}

	switch chainId {
	case ETH_MAINNET,
		ETH_SEPOLIA,
		BSC_MAINNET,
		BSC_TESTNET,
		ARBITRUM_ONE,
		ARBITRUM_NOVA,
		ARBITRUM_SEPOLIA,
		OP_MAINNET,
		OP_SEPOLIA,
		POL_MAINNET,
		POL_TESTNET,
		AVAX_MAINNET,
		AVAX_TESTNET,
		BASE_MAINNET,
		BASE_SEPOLIA:
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
	// case LTC_MAINNET:
	// 	_, err := ltcutil.DecodeAddress(address, &ltcCfg.MainNetParams)
	// 	if err != nil {
	// 		return false
	// 	}
	// 	return true
	// case LTC_TESTNET:
	// 	_, err := ltcutil.DecodeAddress(address, &ltcCfg.TestNet4Params)
	// 	if err != nil {
	// 		return false
	// 	}
	// 	return true
	case TRON_MAINNET, TRON_NILE:
		resultVal, _ := TronValidateAddress(chainId, address)
		return resultVal
	case SOL_MAINNET, SOL_DEVNET:
		_, err := solana.PublicKeyFromBase58(address)
		if err != nil {
			return false
		}

		return true
	case TON_MAINNET, TON_TESTNET:
		resultVal, err := tonAddress.ParseAddr(address)
		if err != nil {
			return false
		}
		if resultVal.Type() == tonAddress.StdAddress {
			return true
		}
		return false
	case XRP_MAINNET, XRP_TESTNET:
		return XrpValidateAddress(address)
	case BCH_MAINNET:
		_, err := bchutil.DecodeAddress(address, &bchCfg.MainNetParams)
		if err != nil {
			return false
		}
		return true
	case BCH_TESTNET:
		_, err := bchutil.DecodeAddress(address, &bchCfg.TestNet3Params)
		if err != nil {
			return false
		}
		return true
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
		ETH_SEPOLIA,
		BSC_MAINNET,
		BSC_TESTNET,
		ARBITRUM_ONE,
		ARBITRUM_NOVA,
		ARBITRUM_SEPOLIA,
		OP_MAINNET,
		OP_SEPOLIA,
		POL_MAINNET,
		POL_TESTNET,
		AVAX_MAINNET,
		AVAX_TESTNET,
		BASE_MAINNET,
		BASE_SEPOLIA:
		client.URL = GetRPCUrlByNetwork(chainId)
		var rpcGeneral response.RPCGeneral
		var jsonRpcRequest request.JsonRpcRequest
		jsonRpcRequest.Id = 1
		jsonRpcRequest.Jsonrpc = "2.0"
		jsonRpcRequest.Method = "eth_getCode"
		jsonRpcRequest.Params = []any{address, "latest"}

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
	case SOL_MAINNET, SOL_DEVNET:
		_, err := solana.PublicKeyFromBase58(address)
		if err != nil {
			return false
		}

		return true
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
		ETH_SEPOLIA,
		BSC_MAINNET,
		BSC_TESTNET,
		ARBITRUM_ONE,
		ARBITRUM_NOVA,
		ARBITRUM_SEPOLIA,
		OP_MAINNET,
		OP_SEPOLIA,
		POL_MAINNET,
		POL_TESTNET,
		AVAX_MAINNET,
		AVAX_TESTNET,
		BASE_MAINNET,
		BASE_SEPOLIA:
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

// func SolanaValidateAddress(rpc, address string) (bool, string) {
// 	client := solanaRpc.NewRpcClient(rpc)

// 	pubKey := solanaCommon.PublicKeyFromString(address)

// 	accountInfo, err := client.GetAccountInfo(context.Background(), pubKey.ToBase58())
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return false, ""
// 	}

// 	if accountInfo.GetResult().Value.Owner == "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
// 		return true, "contract"
// 	} else {
// 		return true, "address"
// 	}
// }

func XrpValidateAddress(address string) bool {
	config := xrpl.ClientConfig{
		URL: "wss://s.altnet.rippletest.net:51233",
	}
	client := xrpl.NewClient(config)
	err := client.Ping([]byte("PING"))
	if err != nil {
		return false
	}

	request := xrpl.BaseRequest{
		"command":      "account_info",
		"account":      address,
		"ledger_index": "validated",
	}

	response, err := client.Request(request)
	if err == nil && response["status"] == "success" {
		return true
	}

	return false
}
