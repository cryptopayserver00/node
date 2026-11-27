package constant

import (
	"fmt"
	"node/global"
	"strings"
)

// GetExplorerAPIBasesByChainId returns ordered etherscan-compatible explorer bases.
// Prefer free Blockscout / Routescan first — free Etherscan V2 keys often only cover ETH.
// Ordinary JSON-RPC nodes cannot list address history — an indexer is required.
func GetExplorerAPIBasesByChainId(chainId uint) []string {
	bases := make([]string, 0, 5)

	switch chainId {
	case ETH_MAINNET:
		bases = append(bases,
			"https://eth.blockscout.com/api",
			"https://api.routescan.io/v2/network/mainnet/evm/1/etherscan/api",
		)
	case ETH_SEPOLIA:
		bases = append(bases,
			"https://eth-sepolia.blockscout.com/api",
			"https://api.routescan.io/v2/network/testnet/evm/11155111/etherscan/api",
		)
	case BSC_MAINNET:
		bases = append(bases,
			"https://api.routescan.io/v2/network/mainnet/evm/56/etherscan/api",
		)
	case BSC_TESTNET:
		// Routescan / free Etherscan V2 often reject chain 97; keep legacy + V2 as best-effort.
		bases = append(bases,
			"https://api.routescan.io/v2/network/testnet/evm/97/etherscan/api",
		)
	case OP_MAINNET:
		bases = append(bases,
			"https://optimism.blockscout.com/api",
			"https://api.routescan.io/v2/network/mainnet/evm/10/etherscan/api",
		)
	case OP_SEPOLIA:
		bases = append(bases,
			"https://optimism-sepolia.blockscout.com/api",
			"https://api.routescan.io/v2/network/testnet/evm/11155420/etherscan/api",
		)
	case ARBITRUM_ONE:
		bases = append(bases,
			"https://arbitrum.blockscout.com/api",
			"https://api.routescan.io/v2/network/mainnet/evm/42161/etherscan/api",
		)
	case ARBITRUM_SEPOLIA:
		bases = append(bases,
			"https://arbitrum-sepolia.blockscout.com/api",
			"https://api.routescan.io/v2/network/testnet/evm/421614/etherscan/api",
		)
	case ARBITRUM_NOVA:
		bases = append(bases,
			"https://nova-explorer.arbitrum.io/api",
			"https://api.routescan.io/v2/network/mainnet/evm/42170/etherscan/api",
		)
	case POL_MAINNET:
		bases = append(bases,
			"https://polygon.blockscout.com/api",
			"https://api.routescan.io/v2/network/mainnet/evm/137/etherscan/api",
		)
	case POL_TESTNET:
		bases = append(bases,
			"https://api.routescan.io/v2/network/testnet/evm/80002/etherscan/api",
		)
	case AVAX_MAINNET:
		bases = append(bases,
			"https://api.routescan.io/v2/network/mainnet/evm/43114/etherscan/api",
			"https://api.snowtrace.io/api",
		)
	case AVAX_TESTNET:
		bases = append(bases,
			"https://api.routescan.io/v2/network/testnet/evm/43113/etherscan/api",
			"https://api-testnet.snowtrace.io/api",
		)
	case BASE_MAINNET:
		bases = append(bases,
			"https://base.blockscout.com/api",
			"https://api.routescan.io/v2/network/mainnet/evm/8453/etherscan/api",
		)
	case BASE_SEPOLIA:
		bases = append(bases,
			"https://base-sepolia.blockscout.com/api",
			"https://api.routescan.io/v2/network/testnet/evm/84532/etherscan/api",
		)
	}

	// Etherscan V2 last: free keys frequently reject non-ETH chains ("upgrade your api plan").
	if etherscanBase := GetEtherscanV2APIBase(chainId); etherscanBase != "" {
		bases = append(bases, etherscanBase)
	}

	return uniqueNonEmpty(bases)
}

// GetExplorerAPIBaseByChainId returns the first configured explorer base (compat).
func GetExplorerAPIBaseByChainId(chainId uint) string {
	bases := GetExplorerAPIBasesByChainId(chainId)
	if len(bases) == 0 {
		return ""
	}
	return bases[0]
}

// GetEtherscanV2APIBase builds the unified Etherscan V2 endpoint when an API key is set.
// Free key: ~3 req/s, 100k/day across mainnets + testnets (incl. Sepolia).
func GetEtherscanV2APIBase(chainId uint) string {
	key := GetRandomEtherscanAPIKey()
	if key == "" {
		return ""
	}
	switch chainId {
	case ETH_MAINNET, ETH_SEPOLIA,
		BSC_MAINNET, BSC_TESTNET,
		OP_MAINNET, OP_SEPOLIA,
		ARBITRUM_ONE, ARBITRUM_SEPOLIA, ARBITRUM_NOVA,
		POL_MAINNET, POL_TESTNET,
		AVAX_MAINNET, AVAX_TESTNET,
		BASE_MAINNET, BASE_SEPOLIA:
		return fmt.Sprintf("https://api.etherscan.io/v2/api?chainid=%d&apikey=%s", chainId, key)
	default:
		return ""
	}
}

func GetRandomEtherscanAPIKey() string {
	return pickCSV(global.NODE_CONFIG.Key.EtherscanApiKey)
}

func GetExplorerTxListURL(chainId uint, address string, page, offset int) string {
	return BuildExplorerTxListURL(GetExplorerAPIBaseByChainId(chainId), address, page, offset)
}

func GetExplorerTokenTxURL(chainId uint, address string, page, offset int) string {
	return BuildExplorerTokenTxURL(GetExplorerAPIBaseByChainId(chainId), address, page, offset)
}

func BuildExplorerTxListURL(base, address string, page, offset int) string {
	if base == "" {
		return ""
	}
	if page <= 0 {
		page = 1
	}
	if offset <= 0 {
		offset = 50
	}
	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	return fmt.Sprintf(
		"%s%smodule=account&action=txlist&address=%s&startblock=0&endblock=99999999&page=%d&offset=%d&sort=desc",
		base, sep, address, page, offset,
	)
}

func BuildExplorerTokenTxURL(base, address string, page, offset int) string {
	if base == "" {
		return ""
	}
	if page <= 0 {
		page = 1
	}
	if offset <= 0 {
		offset = 50
	}
	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	return fmt.Sprintf(
		"%s%smodule=account&action=tokentx&address=%s&page=%d&offset=%d&sort=desc",
		base, sep, address, page, offset,
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

func uniqueNonEmpty(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
