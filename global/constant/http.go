package constant

import (
	"math/rand"
	"node/global"
	"strings"
	"time"
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
)

func GetBscscanUrlByNetwork(network uint) string {
	switch network {
	case BSC_MAINNET:
		return BscscanMainnetAPI
	case BSC_TESTNET:
		return BscscanTestnetAPI
	}

	return ""
}

func GetHttpUrlByNetwork(network uint) string {
	switch network {
	case TRON_MAINNET:
		return TrongridMainnetAPI
	case TRON_NILE:
		return TrongridNileAPI
	}

	return ""
}

func GetBlcokStreamHttpUrlByNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return BlockStreamMainnetAPI
	case BTC_TESTNET:
		return BlockStreamTestnetAPI
	}

	return ""
}

func GetBlockStreamWebsiteTxUrlByNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return BlockStreamWebsiteTxMainnetUrl
	case BTC_TESTNET:
		return BlockStreamWebsiteTxTestnetUrl
	}

	return ""
}

func GetSochainTxUrlByNetwork(network uint) string {
	switch network {
	case LTC_MAINNET:
		return LtcSochainMainnetUrl
	case LTC_TESTNET:
		return LtcSochainTestnetUrl
	}

	return ""
}

func GetRandomHTTPKeyByNetwork(network uint) string {
	rand.Seed(time.Now().UnixNano())

	switch network {
	case TRON_MAINNET:
		index := rand.Intn(len(strings.Split(global.NODE_CONFIG.Key.TrongridMainnetKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.TrongridMainnetKey, ",")[index]
	case TRON_NILE:
		index := rand.Intn(len(strings.Split(global.NODE_CONFIG.Key.TrongridNileKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.TrongridNileKey, ",")[index]
	}

	return ""
}

func GetAllHTTPKeyByNetwork(network uint) []string {
	switch network {
	case TRON_MAINNET:
		return strings.Split(global.NODE_CONFIG.Key.TrongridMainnetKey, ",")
	case TRON_NILE:
		return strings.Split(global.NODE_CONFIG.Key.TrongridNileKey, ",")
	}

	return nil
}

func TronGetBlockByNetwork(network uint) string {
	switch network {
	case TRON_MAINNET:
		return TrongridMainnetAPI + "/walletsolidity/getblock"
	case TRON_NILE:
		return TrongridNileAPI + "/walletsolidity/getblock"
	}

	return ""
}

func TronGetBlockByNumByNetwork(network uint) string {
	switch network {
	case TRON_MAINNET:
		return TrongridMainnetAPI + "/walletsolidity/getblockbynum"
	case TRON_NILE:
		return TrongridNileAPI + "/walletsolidity/getblockbynum"
	}

	return ""
}

func TronGetTxByIdByNetwork(network uint) string {
	switch network {
	case TRON_MAINNET:
		return TrongridMainnetAPI + "/walletsolidity/gettransactionbyid"
	case TRON_NILE:
		return TrongridNileAPI + "/walletsolidity/gettransactionbyid"
	}

	return ""
}

func TronValidateAddressByNetwork(network uint) string {
	switch network {
	case TRON_MAINNET:
		return TrongridMainnetAPI + "/wallet/validateaddress"
	case TRON_NILE:
		return TrongridNileAPI + "/wallet/validateaddress"
	}

	return ""
}

func TronValidateContractAddressByNetwork(network uint) string {
	switch network {
	case TRON_MAINNET:
		return TrongridMainnetAPI + "/wallet/getcontract"
	case TRON_NILE:
		return TrongridNileAPI + "/wallet/getcontract"
	}

	return ""
}
