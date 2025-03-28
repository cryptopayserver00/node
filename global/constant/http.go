package constant

import (
	"math/rand/v2"
	"node/global"
	"strings"
)

var (
	TrongridMainnetAPI = "https://api.trongrid.io"
	TrongridNileAPI    = "https://nile.trongrid.io"

	BlockStreamMainnetAPI = "https://blockstream.info/api"
	BlockStreamTestnetAPI = "https://blockstream.info/testnet/api"

	BlockStreamWebsiteTxMainnetUrl = "https://blockstream.info/tx"
	BlockStreamWebsiteTxTestnetUrl = "https://blockstream.info/testnet/tx"

	LtcSochainMainnetUrl = "https://sochain.com/tx/LTC"
	LtcSochainTestnetUrl = "https://sochain.com/tx/LTCTEST"

	BscscanMainnetAPI = "https://api.bscscan.com"
	BscscanTestnetAPI = "https://api-testnet.bscscan.com"

	TonMainnetAPI = "https://ton.org/global.config.json"
	TonTestnetAPI = "https://ton.org/testnet-global.config.json"

	XRPWsMainnetAPI = "wss://xrplcluster.com"
	XRPWsTestnetAPI = "wss://s.altnet.rippletest.net:51233"
)

func GetBscscanUrlByNetwork(network uint) string {
	switch network {
	case BSC_MAINNET:
		return BscscanMainnetAPI
	case BSC_TESTNET:
		return BscscanTestnetAPI
	default:
		return ""
	}
}

func GetHttpUrlByNetwork(network uint) string {
	switch network {
	case TRON_MAINNET:
		return TrongridMainnetAPI
	case TRON_NILE:
		return TrongridNileAPI
	case TON_MAINNET:
		return TonMainnetAPI
	case TON_TESTNET:
		return TonTestnetAPI
	default:
		return ""
	}
}

func GetBlcokStreamHttpUrlByNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return BlockStreamMainnetAPI
	case BTC_TESTNET:
		return BlockStreamTestnetAPI
	default:
		return ""
	}
}

func GetBlockStreamWebsiteTxUrlByNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return BlockStreamWebsiteTxMainnetUrl
	case BTC_TESTNET:
		return BlockStreamWebsiteTxTestnetUrl
	default:
		return ""
	}
}

func GetSochainTxUrlByNetwork(network uint) string {
	switch network {
	case LTC_MAINNET:
		return LtcSochainMainnetUrl
	case LTC_TESTNET:
		return LtcSochainTestnetUrl
	default:
		return ""
	}

}

func GetRandomHTTPKeyByNetwork(network uint) string {
	switch network {
	case TRON_MAINNET:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.TrongridMainnetKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.TrongridMainnetKey, ",")[index]
	case TRON_NILE:
		index := rand.IntN(len(strings.Split(global.NODE_CONFIG.Key.TrongridNileKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.TrongridNileKey, ",")[index]
	default:
		return ""
	}
}

func GetAllHTTPKeyByNetwork(network uint) []string {
	switch network {
	case TRON_MAINNET:
		return strings.Split(global.NODE_CONFIG.Key.TrongridMainnetKey, ",")
	case TRON_NILE:
		return strings.Split(global.NODE_CONFIG.Key.TrongridNileKey, ",")
	default:
		return nil
	}
}

func TronGetBlockByNetwork(network uint) string {
	switch network {
	case TRON_MAINNET:
		return TrongridMainnetAPI + "/walletsolidity/getblock"
	case TRON_NILE:
		return TrongridNileAPI + "/walletsolidity/getblock"
	default:
		return ""
	}
}

func TronGetBlockByNumByNetwork(network uint) string {
	switch network {
	case TRON_MAINNET:
		return TrongridMainnetAPI + "/walletsolidity/getblockbynum"
	case TRON_NILE:
		return TrongridNileAPI + "/walletsolidity/getblockbynum"
	default:
		return ""
	}
}

func TronGetTxByIdByNetwork(network uint) string {
	switch network {
	case TRON_MAINNET:
		return TrongridMainnetAPI + "/walletsolidity/gettransactionbyid"
	case TRON_NILE:
		return TrongridNileAPI + "/walletsolidity/gettransactionbyid"
	default:
		return ""
	}
}

func TronValidateAddressByNetwork(network uint) string {
	switch network {
	case TRON_MAINNET:
		return TrongridMainnetAPI + "/wallet/validateaddress"
	case TRON_NILE:
		return TrongridNileAPI + "/wallet/validateaddress"
	default:
		return ""
	}
}

func TronValidateContractAddressByNetwork(network uint) string {
	switch network {
	case TRON_MAINNET:
		return TrongridMainnetAPI + "/wallet/getcontract"
	case TRON_NILE:
		return TrongridNileAPI + "/wallet/getcontract"
	default:
		return ""
	}
}

func XrpWsByNetwork(network uint) string {
	switch network {
	case XRP_MAINNET:
		return XRPWsMainnetAPI
	case XRP_TESTNET:
		return XRPWsTestnetAPI
	default:
		return ""
	}
}
