package constant

import "fmt"

// Free etherscan-compatible explorers (Blockscout / Routescan).
// No API key required for basic account tx queries.
func GetExplorerAPIBaseByChainId(chainId uint) string {
	switch chainId {
	case ETH_MAINNET:
		return "https://eth.blockscout.com/api"
	case ETH_SEPOLIA:
		return "https://eth-sepolia.blockscout.com/api"
	case BSC_MAINNET:
		return "https://api.routescan.io/v2/network/mainnet/evm/56/etherscan/api"
	case BSC_TESTNET:
		return "https://api.routescan.io/v2/network/testnet/evm/97/etherscan/api"
	case OP_MAINNET:
		return "https://optimism.blockscout.com/api"
	case OP_SEPOLIA:
		return "https://optimism-sepolia.blockscout.com/api"
	case ARBITRUM_ONE:
		return "https://arbitrum.blockscout.com/api"
	case ARBITRUM_SEPOLIA:
		return "https://arbitrum-sepolia.blockscout.com/api"
	case ARBITRUM_NOVA:
		return "https://api.routescan.io/v2/network/mainnet/evm/42170/etherscan/api"
	case POL_MAINNET:
		return "https://polygon.blockscout.com/api"
	case POL_TESTNET:
		return "https://api.routescan.io/v2/network/testnet/evm/80002/etherscan/api"
	case AVAX_MAINNET:
		return "https://api.routescan.io/v2/network/mainnet/evm/43114/etherscan/api"
	case AVAX_TESTNET:
		return "https://api.routescan.io/v2/network/testnet/evm/43113/etherscan/api"
	case BASE_MAINNET:
		return "https://base.blockscout.com/api"
	case BASE_SEPOLIA:
		return "https://base-sepolia.blockscout.com/api"
	default:
		return ""
	}
}

func GetExplorerTxListURL(chainId uint, address string, page, offset int) string {
	base := GetExplorerAPIBaseByChainId(chainId)
	if base == "" {
		return ""
	}
	if page <= 0 {
		page = 1
	}
	if offset <= 0 {
		offset = 50
	}
	return fmt.Sprintf(
		"%s?module=account&action=txlist&address=%s&startblock=0&endblock=99999999&page=%d&offset=%d&sort=desc",
		base, address, page, offset,
	)
}

func GetExplorerTokenTxURL(chainId uint, address string, page, offset int) string {
	base := GetExplorerAPIBaseByChainId(chainId)
	if base == "" {
		return ""
	}
	if page <= 0 {
		page = 1
	}
	if offset <= 0 {
		offset = 50
	}
	return fmt.Sprintf(
		"%s?module=account&action=tokentx&address=%s&page=%d&offset=%d&sort=desc",
		base, address, page, offset,
	)
}

// GetNativeAssetByChainId returns the wallet display symbol for native gas token.
func GetNativeAssetByChainId(chainId uint) string {
	switch chainId {
	case ETH_MAINNET, ETH_SEPOLIA, OP_MAINNET, OP_SEPOLIA, ARBITRUM_ONE, ARBITRUM_NOVA, ARBITRUM_SEPOLIA, BASE_MAINNET, BASE_SEPOLIA:
		return ETH
	case BSC_MAINNET, BSC_TESTNET:
		return BNB
	case POL_MAINNET, POL_TESTNET:
		return POL
	case AVAX_MAINNET, AVAX_TESTNET:
		return AVAX
	default:
		return ETH
	}
}

func GetTonCenterAPIBase(chainId uint) string {
	switch chainId {
	case TON_MAINNET:
		return "https://toncenter.com/api/v2"
	case TON_TESTNET:
		return "https://testnet.toncenter.com/api/v2"
	default:
		return ""
	}
}
