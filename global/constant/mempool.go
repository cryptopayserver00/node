package constant

func MempoolGetBlockHeightByNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return "https://mempool.space/api/blocks/tip/height"
	case BTC_TESTNET:
		return "https://mempool.space/testnet/api/blocks/tip/height"
	case LTC_MAINNET:
		return "https://litecoinspace.org/api/blocks/tip/height"
	}

	return ""
}

func MempoolGetBlockTransactionByNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return "https://mempool.space/api/block/%s/txs/%d"
	case BTC_TESTNET:
		return "https://mempool.space/testnet/api/block/%s/txs/%d"
	case LTC_MAINNET:
		return "https://litecoinspace.org/api/block/%s/txs/%d"
	}

	return ""
}

func MempoolGetBlockHashByNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return "https://mempool.space/api/block-height/%d"
	case BTC_TESTNET:
		return "https://mempool.space/testnet/api/block-height/%d"
	case LTC_MAINNET:
		return "https://litecoinspace.org/api/block-height/%d"
	}

	return ""
}

func MempoolGetBlockByNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return "https://mempool.space/api/block/%s"
	case BTC_TESTNET:
		return "https://mempool.space/testnet/api/block/%s"
	case LTC_MAINNET:
		return "https://litecoinspace.org/api/block/%s"
	}

	return ""
}

func MempoolGetTransctionByNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return "https://mempool.space/api/tx/%s"
	case BTC_TESTNET:
		return "https://mempool.space/testnet/api/tx/%s"
	case LTC_MAINNET:
		return "https://litecoinspace.org/api/tx/%s"
	}

	return ""
}

func MempoolGetUtxoByNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return "https://mempool.space/api/address/%s/utxo"
	case BTC_TESTNET:
		return "https://mempool.space/testnet/api/address/%s/utxo"
	case LTC_MAINNET:
		return "https://litecoinspace.org/api/address/%s/utxo"
	case LTC_TESTNET:
		return "https://litecoinspace.org/testnet/api/address/%s/utxo"
	}

	return ""
}

func MempoolGetFeesyNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return "https://mempool.space/api/v1/fees/recommended"
	case BTC_TESTNET:
		return "https://mempool.space/testnet/api/v1/fees/recommended"
	case LTC_MAINNET:
		return "https://litecoinspace.org/api/v1/fees/recommended"
	case LTC_TESTNET:
		return "https://litecoinspace.org/testnet/api/v1/fees/recommended"
	}

	return ""
}

func MempoolBroadcastByNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return "https://mempool.space/api/tx"
	case BTC_TESTNET:
		return "https://mempool.space/testnet/api/tx"
	case LTC_MAINNET:
		return "https://litecoinspace.org/api/tx"
	case LTC_TESTNET:
		return "https://litecoinspace.org/testnet/api/tx"
	}

	return ""
}
