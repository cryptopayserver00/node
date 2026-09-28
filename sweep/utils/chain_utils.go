package utils

import (
	"fmt"
	"node/global"
	"node/global/constant"
	"node/model"
	"node/utils"
	"slices"
	"strings"
	"sync"
)

// ---------------------------------------------------------------------
// 启动期构建一次的地址索引，避免每笔交易都做 O(chains × coins) 的线性扫描 + 重复归一化
// ---------------------------------------------------------------------

// nativeCoinKey 是 BTC/LTC/BCH 这类「按主币标记，不按地址区分」的链使用的哨兵 key。
const nativeCoinKey = "__native_coin__"

var (
	contractIndexOnce sync.Once
	// contractIndex: chainId -> 归一化后的地址(或 nativeCoinKey) -> Coin
	contractIndex map[uint]map[string]model.Coin
)

// normalizeContractKey 按链类型把地址转换成统一的比较 key。
// 新增一条链时，如果这里没有对应分支，会走 default 并打日志，方便及时发现遗漏，
// 而不是让 GetContractInfo 静默返回“不支持”。
func normalizeContractKey(chainId uint, contractAddress string, isMainCoin bool) (string, bool) {
	switch chainId {
	case constant.ETH_MAINNET,
		constant.ETH_SEPOLIA,
		constant.BSC_MAINNET,
		constant.BSC_TESTNET,
		constant.OP_MAINNET,
		constant.OP_SEPOLIA,
		constant.ARBITRUM_ONE,
		constant.ARBITRUM_NOVA,
		constant.ARBITRUM_SEPOLIA,
		constant.POL_MAINNET,
		constant.POL_TESTNET,
		constant.AVAX_MAINNET,
		constant.AVAX_TESTNET,
		constant.BASE_MAINNET,
		constant.BASE_SEPOLIA:
		return utils.HexToAddress(contractAddress), true

	case constant.TRON_NILE, constant.TRON_MAINNET,
		constant.SOL_MAINNET, constant.SOL_DEVNET,
		constant.XRP_MAINNET, constant.XRP_TESTNET:
		return strings.ToLower(contractAddress), true

	case constant.BTC_MAINNET, constant.BTC_TESTNET,
		constant.LTC_MAINNET, constant.LTC_TESTNET,
		constant.BCH_MAINNET, constant.BCH_TESTNET:
		if isMainCoin {
			return nativeCoinKey, true
		}
		return "", false
	case constant.TON_MAINNET, constant.TON_TESTNET:
		return contractAddress, true

	default:
		// 出现这个日志，说明某条链已经加进了 JoinSweep，但这里还没配对应的地址归一化规则。
		global.NODE_LOG.Error(fmt.Sprintf("chain %d is not handled in normalizeContractKey, GetContractInfo will always miss for it", chainId))
		return "", false
	}
}

func buildContractIndex() {
	contractIndex = make(map[uint]map[string]model.Coin, len(model.ChainList))

	for _, element := range model.ChainList {
		coinMap := make(map[string]model.Coin, len(element.Coins))

		for _, coin := range element.Coins {
			key, ok := normalizeContractKey(element.ChainId, coin.Contract, coin.IsMainCoin)
			if !ok {
				continue
			}
			coinMap[key] = coin
		}
		contractIndex[element.ChainId] = coinMap
	}
}

func IsChainJoinSweep(chainId uint) bool {
	if chainId == 0 {
		return false
	}
	return slices.Contains(constant.JoinSweep, chainId)
}

func GetContractInfo(chainId uint, contractAddress string) (bool, string, string, int) {
	if !IsChainJoinSweep(chainId) {
		return false, "", "", 0
	}

	contractIndexOnce.Do(buildContractIndex)

	coins, ok := contractIndex[chainId]
	if !ok {
		return false, "", "", 0
	}

	// BTC/LTC/BCH 这类主币场景，直接用哨兵 key 命中。
	if coin, found := coins[nativeCoinKey]; found {
		return true, coin.Symbol, coin.Contract, coin.Decimals
	}

	key, ok := normalizeContractKey(chainId, contractAddress, false)
	if !ok {
		return false, "", "", 0
	}

	if coin, found := coins[key]; found {
		return true, coin.Symbol, coin.Contract, coin.Decimals
	}

	return false, "", "", 0
}

func GetContractInfoByChainIdAndSymbol(chainId uint, symbol string) (bool, string, string, int) {
	if !IsChainJoinSweep(chainId) {
		return false, "", "", 0
	}

	for _, element := range model.ChainList {
		if element.ChainId != chainId {
			continue
		}

		for _, coin := range element.Coins {
			if coin.Symbol == symbol {
				return true, coin.Symbol, coin.Contract, coin.Decimals
			}
		}
	}
	return false, "", "", 0
}

func GetCoinsByChainId(chainId uint) (bool, []model.Coin) {
	if !IsChainJoinSweep(chainId) {
		return false, nil
	}

	for _, element := range model.ChainList {
		if element.ChainId == chainId {
			return true, element.Coins
		}
	}
	return false, nil
}

func IsFreeCoinSupport(chainId uint, freeCoin string) bool {
	if !constant.IsTestnetSupport(chainId) {
		return false
	}

	for _, element := range model.ChainList {
		if element.ChainId != chainId {
			continue
		}

		for _, coin := range element.Coins {
			if strings.EqualFold(coin.Symbol, freeCoin) {
				return true
			}
		}
	}

	return false
}
