package utils

import (
	"node/global/constant"
	"node/model"
	"node/utils"
	"strings"
)

func IsChainJoinSweep(chainId uint) bool {
	if chainId == 0 {
		return false
	}

	for _, v := range constant.JoinSweep {
		if v == chainId {
			return true
		}
	}

	return false
}

// isContract, symbol, contractAddress, decimals
func GetContractInfo(chainId uint, contractAddress string) (bool, string, string, int) {
	if !IsChainJoinSweep(chainId) {
		return false, "", "", 0
	}

	for _, element := range model.ChainList {
		if element.ChainId != chainId {
			continue
		}

		for _, coin := range element.Coins {
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
				if utils.HexToAddress(coin.Contract) == utils.HexToAddress(contractAddress) {
					return true, coin.Symbol, coin.Contract, coin.Decimals
				}
			case constant.TRON_NILE, constant.TRON_MAINNET, constant.SOL_MAINNET, constant.SOL_DEVNET, constant.XRP_MAINNET, constant.XRP_TESTNET:
				if strings.EqualFold(coin.Contract, contractAddress) {
					return true, coin.Symbol, coin.Contract, coin.Decimals
				}
			case constant.BTC_MAINNET, constant.BTC_TESTNET:
				if coin.IsMainCoin {
					return true, coin.Symbol, coin.Contract, coin.Decimals
				}
			case constant.LTC_MAINNET, constant.LTC_TESTNET:
				if coin.IsMainCoin {
					return true, coin.Symbol, coin.Contract, coin.Decimals
				}
			}
		}

		return false, "", "", 0

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
			if coin.Symbol == freeCoin {
				return true
			}
		}
	}

	return false
}
