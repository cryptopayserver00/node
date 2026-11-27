package constant

func MempoolGetBlockHeightByNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return "https://mempool.space/api/blocks/tip/height"
	case BTC_TESTNET:
		return "https://mempool.space/testnet/api/blocks/tip/height"
	case LTC_MAINNET:
		return "https://litecoinspace.org/api/blocks/tip/height"
	case BCH_MAINNET:
		return "https://bch.fullstack.cash/v6/full-node/blockchain/getBlockCount"
	default:
		panic("not support")
	}
}

func MempoolGetBlockTransactionByNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return "https://mempool.space/api/block/%s/txs/%d"
	case BTC_TESTNET:
		return "https://mempool.space/testnet/api/block/%s/txs/%d"
	case LTC_MAINNET:
		return "https://litecoinspace.org/api/block/%s/txs/%d"
	case BCH_MAINNET:
		return "https://bchexplorer.cash/api/block/%s/txs/%d"
	default:
		panic("not support")
	}
}

func MempoolGetBlockHashByNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return "https://mempool.space/api/block-height/%d"
	case BTC_TESTNET:
		return "https://mempool.space/testnet/api/block-height/%d"
	case LTC_MAINNET:
		return "https://litecoinspace.org/api/block-height/%d"
	case BCH_MAINNET:
		return "https://bchexplorer.cash/api/block-height/%d"
	default:
		panic("not support")
	}
}

func MempoolGetBlockByNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return "https://mempool.space/api/block/%s"
	case BTC_TESTNET:
		return "https://mempool.space/testnet/api/block/%s"
	case LTC_MAINNET:
		return "https://litecoinspace.org/api/block/%s"
	case BCH_MAINNET:
		return "https://bchexplorer.cash/api/block/%s"
	default:
		panic("not support")
	}
}

func MempoolGetTransctionByNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return "https://mempool.space/api/tx/%s"
	case BTC_TESTNET:
		return "https://mempool.space/testnet/api/tx/%s"
	case LTC_MAINNET:
		return "https://litecoinspace.org/api/tx/%s"
	case BCH_MAINNET:
		return "https://bchexplorer.cash/api/tx/%s"
	default:
		panic("not support")
	}
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
	case BCH_MAINNET:
		return "https://bchexplorer.cash/api/address/%s/utxo"
	default:
		panic("not support")
	}
}

// MempoolGetAddressTxsByNetwork returns confirmed+recent txs for an address (Esplora).
func MempoolGetAddressTxsByNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return "https://mempool.space/api/address/%s/txs"
	case BTC_TESTNET:
		return "https://mempool.space/testnet/api/address/%s/txs"
	case LTC_MAINNET:
		return "https://litecoinspace.org/api/address/%s/txs"
	case LTC_TESTNET:
		return "https://litecoinspace.org/testnet/api/address/%s/txs"
	case BCH_MAINNET:
		return "https://bchexplorer.cash/api/address/%s/txs"
	default:
		return ""
	}
}

// MempoolGetAddressMempoolTxsByNetwork returns unconfirmed txs for an address.
func MempoolGetAddressMempoolTxsByNetwork(network uint) string {
	switch network {
	case BTC_MAINNET:
		return "https://mempool.space/api/address/%s/txs/mempool"
	case BTC_TESTNET:
		return "https://mempool.space/testnet/api/address/%s/txs/mempool"
	case LTC_MAINNET:
		return "https://litecoinspace.org/api/address/%s/txs/mempool"
	case LTC_TESTNET:
		return "https://litecoinspace.org/testnet/api/address/%s/txs/mempool"
	case BCH_MAINNET:
		return "https://bchexplorer.cash/api/address/%s/txs/mempool"
	default:
		return ""
	}
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
	case BCH_MAINNET:
		return "https://bchexplorer.cash/api/v1/fees/recommended"
	default:
		panic("not support")
	}
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
	case BCH_MAINNET:
		return "https://bchexplorer.cash/api/tx"
	default:
		panic("not support")
	}
}
