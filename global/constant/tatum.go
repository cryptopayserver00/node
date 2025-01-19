package constant

import (
	"math/rand"
	"node/global"
	"strings"
	"time"
)

var (
	TatumAPI = "https://api.tatum.io/v3"

	// Bitcoin
	TatumGetBitcoinInfo                = TatumAPI + "/bitcoin/info"
	TatumGetBitcoinBlockByHashOrHeight = TatumAPI + "/bitcoin/block/"
	TatumGetBitcoinTxByHash            = TatumAPI + "/bitcoin/transaction/"
	TatumGetBitcoinBalance             = TatumAPI + "/bitcoin/address/balance/"
	TatumGetBitcoinFeeRate             = TatumAPI + "/blockchain/fee/BTC"
	TatumGetBitcoinBroadcast           = TatumAPI + "/bitcoin/broadcast"
	TatumGetBitcoinTransactions        = TatumAPI + "/bitcoin/transaction/address/"

	// Litecoin
	TatumGetLitecoinInfo                = TatumAPI + "/litecoin/info"
	TatumGetLitecoinBlockByHashOrHeight = TatumAPI + "/litecoin/block/"
	TatumGetLitecoinTxByHash            = TatumAPI + "/litecoin/transaction/"
	TatumGetLitecoinBalance             = TatumAPI + "/litecoin/address/balance/"
	TatumGetLitecoinFeeRate             = TatumAPI + "/blockchain/fee/LTC"
	TatumGetLitecoinBroadcast           = TatumAPI + "/litecoin/broadcast"
	TatumGetLitecoinTransactions        = TatumAPI + "/litecoin/transaction/address/"
	TatumGetLitecoinUtxo                = TatumAPI + "/data/utxos"
)

var TatumSupportChain = []uint{
	BTC_MAINNET,
	BTC_TESTNET,
	ETH_MAINNET,
	ETH_SEPOLIA,
	LTC_MAINNET,
	LTC_TESTNET,
	BSC_MAINNET,
	BSC_TESTNET,
	TRON_MAINNET,
}

func IsNetworkSupportTatum(id uint) bool {
	for _, v := range TatumSupportChain {
		if v == id {
			return true
		}
	}

	return false
}

func GetTatumRandomKeyByNetwork(id uint) string {
	rand.Seed(time.Now().UnixNano())

	switch id {
	case BTC_MAINNET, ETH_MAINNET, LTC_MAINNET, BSC_MAINNET, TRON_MAINNET:
		index := rand.Intn(len(strings.Split(global.NODE_CONFIG.Key.TatumMainnetKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.TatumMainnetKey, ",")[index]
	case BTC_TESTNET, ETH_SEPOLIA, LTC_TESTNET, BSC_TESTNET:
		index := rand.Intn(len(strings.Split(global.NODE_CONFIG.Key.TatumTestnetKey, ",")))
		return strings.Split(global.NODE_CONFIG.Key.TatumTestnetKey, ",")[index]
	}

	return ""
}

func GetAllTatumAPiKey(id uint) []string {
	switch id {
	case BTC_MAINNET, ETH_MAINNET, LTC_MAINNET, BSC_MAINNET, TRON_MAINNET:
		return strings.Split(global.NODE_CONFIG.Key.TatumMainnetKey, ",")
	case BTC_TESTNET, ETH_SEPOLIA, LTC_TESTNET, BSC_TESTNET:
		return strings.Split(global.NODE_CONFIG.Key.TatumTestnetKey, ",")

	}
	return nil
}
