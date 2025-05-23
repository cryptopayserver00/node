package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/model"
	"node/model/node/request"
	"node/model/node/response"

	"gorm.io/gorm"
)

func (n *NService) InitChainList() (err error) {
	result := global.NODE_DB.Find(&model.Chain{})
	if result.Error != nil {
		return result.Error
	}

	if !(result.RowsAffected == 0) {
		global.NODE_LOG.Info("chain already initialize")

		return n.UpdateChainListFromDB()
	}

	err = constant.UpdateChainListFromFile()
	if err != nil {
		global.NODE_LOG.Info(err.Error())
		return
	}

	if len(model.ChainList) > 0 {
		for _, element := range model.ChainList {
			if len(element.Coins) == 0 {
				continue
			}

			for _, coin := range element.Coins {
				var chainModel model.Chain
				chainModel.Name = element.Name
				chainModel.Chain = element.Chain
				chainModel.ChainId = element.ChainId
				chainModel.NetworkId = element.NetworkId
				chainModel.Symbol = coin.Symbol
				chainModel.Decimals = coin.Decimals
				chainModel.Contract = coin.Contract
				chainModel.IsMainCoin = coin.IsMainCoin
				chainModel.Status = 1

				if err = global.NODE_DB.Create(&chainModel).Error; err != nil {
					return
				}
			}

		}
	}
	return nil
}

func (n *NService) UpdateChainListFromDB() (err error) {
	var chains []model.Chain
	err = global.NODE_DB.Find(&chains).Error
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	if len(chains) > 0 {
		var infos []model.ChainInfo
		var ethMainnetChain, ethSepoliaChain, btcMainnetChain, btcTestnetChain, ltcMainnetChain, ltcTestnetChain, bscMainnetChain, bscTestnetChain, opMainnetChain, opSepoliaChain, arbOneChain, arbNovaChain, arbSepoliaChain, tronMainnetChain, tronNileChain, solMainnetChain, solDevnetChain, tonMainnetChain, tonTestnetChain, xrpMainnetChain, xrpTestnetChain, bchMainnetChain, bchTestnetChain, polMainnetChain, polTestnetChain, avaxMainnetChain, avaxTestnetChain, baseMainnetChain, baseSepoliaChain model.ChainInfo
		for _, v := range chains {
			var coin model.Coin

			coin.Symbol = v.Symbol
			coin.Decimals = v.Decimals
			coin.Contract = v.Contract
			coin.IsMainCoin = v.IsMainCoin

			switch v.ChainId {
			case constant.ETH_MAINNET:
				if ethMainnetChain.Name == "" {
					ethMainnetChain.Name = v.Name
					ethMainnetChain.Chain = v.Chain
					ethMainnetChain.ChainId = v.ChainId
					ethMainnetChain.NetworkId = v.NetworkId
				}

				ethMainnetChain.Coins = append(ethMainnetChain.Coins, coin)
			case constant.ETH_SEPOLIA:
				if ethSepoliaChain.Name == "" {
					ethSepoliaChain.Name = v.Name
					ethSepoliaChain.Chain = v.Chain
					ethSepoliaChain.ChainId = v.ChainId
					ethSepoliaChain.NetworkId = v.NetworkId
				}
				ethSepoliaChain.Coins = append(ethSepoliaChain.Coins, coin)
			case constant.BSC_MAINNET:
				if bscMainnetChain.Name == "" {
					bscMainnetChain.Name = v.Name
					bscMainnetChain.Chain = v.Chain
					bscMainnetChain.ChainId = v.ChainId
					bscMainnetChain.NetworkId = v.NetworkId
				}

				bscMainnetChain.Coins = append(bscMainnetChain.Coins, coin)
			case constant.BSC_TESTNET:
				if bscTestnetChain.Name == "" {
					bscTestnetChain.Name = v.Name
					bscTestnetChain.Chain = v.Chain
					bscTestnetChain.ChainId = v.ChainId
					bscTestnetChain.NetworkId = v.NetworkId
				}

				bscTestnetChain.Coins = append(bscTestnetChain.Coins, coin)
			case constant.BTC_MAINNET:
				if btcMainnetChain.Name == "" {
					btcMainnetChain.Name = v.Name
					btcMainnetChain.Chain = v.Chain
					btcMainnetChain.ChainId = v.ChainId
					btcMainnetChain.NetworkId = v.NetworkId
				}

				btcMainnetChain.Coins = append(btcMainnetChain.Coins, coin)
			case constant.BTC_TESTNET:
				if btcTestnetChain.Name == "" {
					btcTestnetChain.Name = v.Name
					btcTestnetChain.Chain = v.Chain
					btcTestnetChain.ChainId = v.ChainId
					btcTestnetChain.NetworkId = v.NetworkId
				}

				btcTestnetChain.Coins = append(btcTestnetChain.Coins, coin)
			case constant.LTC_MAINNET:
				if ltcMainnetChain.Name == "" {
					ltcMainnetChain.Name = v.Name
					ltcMainnetChain.Chain = v.Chain
					ltcMainnetChain.ChainId = v.ChainId
					ltcMainnetChain.NetworkId = v.NetworkId
				}

				ltcMainnetChain.Coins = append(ltcMainnetChain.Coins, coin)
			case constant.LTC_TESTNET:
				if ltcTestnetChain.Name == "" {
					ltcTestnetChain.Name = v.Name
					ltcTestnetChain.Chain = v.Chain
					ltcTestnetChain.ChainId = v.ChainId
					ltcTestnetChain.NetworkId = v.NetworkId
				}

				ltcTestnetChain.Coins = append(ltcTestnetChain.Coins, coin)
			case constant.OP_MAINNET:
				if opMainnetChain.Name == "" {
					opMainnetChain.Name = v.Name
					opMainnetChain.Chain = v.Chain
					opMainnetChain.ChainId = v.ChainId
					opMainnetChain.NetworkId = v.NetworkId
				}

				opMainnetChain.Coins = append(opMainnetChain.Coins, coin)
			case constant.OP_SEPOLIA:
				if opSepoliaChain.Name == "" {
					opSepoliaChain.Name = v.Name
					opSepoliaChain.Chain = v.Chain
					opSepoliaChain.ChainId = v.ChainId
					opSepoliaChain.NetworkId = v.NetworkId
				}

				opSepoliaChain.Coins = append(opSepoliaChain.Coins, coin)
			case constant.ARBITRUM_ONE:
				if arbOneChain.Name == "" {
					arbOneChain.Name = v.Name
					arbOneChain.Chain = v.Chain
					arbOneChain.ChainId = v.ChainId
					arbOneChain.NetworkId = v.NetworkId
				}

				arbOneChain.Coins = append(arbOneChain.Coins, coin)
			case constant.ARBITRUM_NOVA:
				if arbNovaChain.Name == "" {
					arbNovaChain.Name = v.Name
					arbNovaChain.Chain = v.Chain
					arbNovaChain.ChainId = v.ChainId
					arbNovaChain.NetworkId = v.NetworkId
				}

				arbNovaChain.Coins = append(arbNovaChain.Coins, coin)
			case constant.ARBITRUM_SEPOLIA:
				if arbSepoliaChain.Name == "" {
					arbSepoliaChain.Name = v.Name
					arbSepoliaChain.Chain = v.Chain
					arbSepoliaChain.ChainId = v.ChainId
					arbSepoliaChain.NetworkId = v.NetworkId
				}

				arbSepoliaChain.Coins = append(arbSepoliaChain.Coins, coin)
			case constant.TRON_MAINNET:
				if tronMainnetChain.Name == "" {
					tronMainnetChain.Name = v.Name
					tronMainnetChain.Chain = v.Chain
					tronMainnetChain.ChainId = v.ChainId
					tronMainnetChain.NetworkId = v.NetworkId
				}

				tronMainnetChain.Coins = append(tronMainnetChain.Coins, coin)
			case constant.TRON_NILE:
				if tronNileChain.Name == "" {
					tronNileChain.Name = v.Name
					tronNileChain.Chain = v.Chain
					tronNileChain.ChainId = v.ChainId
					tronNileChain.NetworkId = v.NetworkId
				}

				tronNileChain.Coins = append(tronNileChain.Coins, coin)
			case constant.SOL_MAINNET:
				if solMainnetChain.Name == "" {
					solMainnetChain.Name = v.Name
					solMainnetChain.Chain = v.Chain
					solMainnetChain.ChainId = v.ChainId
					solMainnetChain.NetworkId = v.NetworkId
				}

				solMainnetChain.Coins = append(solMainnetChain.Coins, coin)
			case constant.SOL_DEVNET:
				if solDevnetChain.Name == "" {
					solDevnetChain.Name = v.Name
					solDevnetChain.Chain = v.Chain
					solDevnetChain.ChainId = v.ChainId
					solDevnetChain.NetworkId = v.NetworkId
				}

				solDevnetChain.Coins = append(solDevnetChain.Coins, coin)
			case constant.TON_MAINNET:
				if tonMainnetChain.Name == "" {
					tonMainnetChain.Name = v.Name
					tonMainnetChain.Chain = v.Chain
					tonMainnetChain.ChainId = v.ChainId
					tonMainnetChain.NetworkId = v.NetworkId
				}

				tonMainnetChain.Coins = append(tonMainnetChain.Coins, coin)
			case constant.TON_TESTNET:
				if tonTestnetChain.Name == "" {
					tonTestnetChain.Name = v.Name
					tonTestnetChain.Chain = v.Chain
					tonTestnetChain.ChainId = v.ChainId
					tonTestnetChain.NetworkId = v.NetworkId
				}

				tonTestnetChain.Coins = append(tonTestnetChain.Coins, coin)
			case constant.XRP_MAINNET:
				if xrpMainnetChain.Name == "" {
					xrpMainnetChain.Name = v.Name
					xrpMainnetChain.Chain = v.Chain
					xrpMainnetChain.ChainId = v.ChainId
					xrpMainnetChain.NetworkId = v.NetworkId
				}

				xrpMainnetChain.Coins = append(xrpMainnetChain.Coins, coin)
			case constant.XRP_TESTNET:
				if xrpTestnetChain.Name == "" {
					xrpTestnetChain.Name = v.Name
					xrpTestnetChain.Chain = v.Chain
					xrpTestnetChain.ChainId = v.ChainId
					xrpTestnetChain.NetworkId = v.NetworkId
				}

				xrpTestnetChain.Coins = append(xrpTestnetChain.Coins, coin)
			case constant.BCH_MAINNET:
				if bchMainnetChain.Name == "" {
					bchMainnetChain.Name = v.Name
					bchMainnetChain.Chain = v.Chain
					bchMainnetChain.ChainId = v.ChainId
					bchMainnetChain.NetworkId = v.NetworkId
				}

				bchMainnetChain.Coins = append(bchMainnetChain.Coins, coin)
			case constant.BCH_TESTNET:
				if bchTestnetChain.Name == "" {
					bchTestnetChain.Name = v.Name
					bchTestnetChain.Chain = v.Chain
					bchTestnetChain.ChainId = v.ChainId
					bchTestnetChain.NetworkId = v.NetworkId
				}

				bchTestnetChain.Coins = append(bchTestnetChain.Coins, coin)
			case constant.POL_MAINNET:
				if polMainnetChain.Name == "" {
					polMainnetChain.Name = v.Name
					polMainnetChain.Chain = v.Chain
					polMainnetChain.ChainId = v.ChainId
					polMainnetChain.NetworkId = v.NetworkId
				}

				polMainnetChain.Coins = append(polMainnetChain.Coins, coin)
			case constant.POL_TESTNET:
				if polTestnetChain.Name == "" {
					polTestnetChain.Name = v.Name
					polTestnetChain.Chain = v.Chain
					polTestnetChain.ChainId = v.ChainId
					polTestnetChain.NetworkId = v.NetworkId
				}

				polTestnetChain.Coins = append(polTestnetChain.Coins, coin)
			case constant.AVAX_MAINNET:
				if avaxMainnetChain.Name == "" {
					avaxMainnetChain.Name = v.Name
					avaxMainnetChain.Chain = v.Chain
					avaxMainnetChain.ChainId = v.ChainId
					avaxMainnetChain.NetworkId = v.NetworkId
				}

				avaxMainnetChain.Coins = append(avaxMainnetChain.Coins, coin)
			case constant.AVAX_TESTNET:
				if avaxTestnetChain.Name == "" {
					avaxTestnetChain.Name = v.Name
					avaxTestnetChain.Chain = v.Chain
					avaxTestnetChain.ChainId = v.ChainId
					avaxTestnetChain.NetworkId = v.NetworkId
				}

				avaxTestnetChain.Coins = append(avaxTestnetChain.Coins, coin)
			case constant.BASE_MAINNET:
				if baseMainnetChain.Name == "" {
					baseMainnetChain.Name = v.Name
					baseMainnetChain.Chain = v.Chain
					baseMainnetChain.ChainId = v.ChainId
					baseMainnetChain.NetworkId = v.NetworkId
				}

				baseMainnetChain.Coins = append(baseMainnetChain.Coins, coin)
			case constant.BASE_SEPOLIA:
				if baseSepoliaChain.Name == "" {
					baseSepoliaChain.Name = v.Name
					baseSepoliaChain.Chain = v.Chain
					baseSepoliaChain.ChainId = v.ChainId
					baseSepoliaChain.NetworkId = v.NetworkId
				}

				baseSepoliaChain.Coins = append(baseSepoliaChain.Coins, coin)
			}
		}

		infos = append(infos, ethMainnetChain, ethSepoliaChain, bscMainnetChain, bscTestnetChain, opMainnetChain, opSepoliaChain, arbOneChain, arbNovaChain, arbSepoliaChain, tronMainnetChain, tronNileChain, btcMainnetChain, btcTestnetChain, ltcTestnetChain, ltcMainnetChain, solMainnetChain, solDevnetChain, tonMainnetChain, tonTestnetChain, xrpMainnetChain, xrpTestnetChain, bchMainnetChain, bchTestnetChain, polMainnetChain, polTestnetChain, avaxMainnetChain, avaxTestnetChain, baseMainnetChain, baseSepoliaChain)
		model.ChainList = infos
	}

	return nil
}

func (n *NService) StoreChainContract(chain request.StoreChainContract) (err error) {
	return n.saveChainContract(chain.ChainId, chain.Decimals, chain.Contract, chain.Symbol)
}

func (n *NService) BulkStorageChainContract(contracts request.BulkStoreChainContract) (errChainResponses response.BulkStoreChainContractResponse, err error) {
	if len(contracts.BulkStorage) > 0 {

		for _, v := range contracts.BulkStorage {
			if err = n.saveChainContract(v.ChainId, v.Decimals, v.Contract, v.Symbol); err != nil {
				var errChain response.StoreChainContract
				errChain.ChainId = v.ChainId
				errChain.Contract = v.Contract
				errChainResponses.BulkStorage = append(errChainResponses.BulkStorage, errChain)
				global.NODE_LOG.Error(err.Error())

				continue
			}
		}

		if len(errChainResponses.BulkStorage) > 0 {
			return errChainResponses, errors.New("some contracts failed to store")
		}
	}

	return
}

func (n *NService) GetChainByChainIdAndContractAddress(chainId uint, contractAddress string) (hasChain bool, err error) {
	var findChain model.Chain

	err = global.NODE_DB.Where("chain_id = ? AND contract = ?", chainId, contractAddress).First(&findChain).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	if findChain.ID > 0 {
		return true, nil
	}

	return false, nil
}

func (n *NService) saveChainContract(chainId uint, decimals int, contract, symbol string) (err error) {
	if !constant.IsNetworkSupport(chainId) {
		return errors.New("do not support network")
	}

	contract = constant.AddressToLower(chainId, contract)

	if !constant.IsAddressContractSupport(chainId, contract) {
		return fmt.Errorf("do not support contract address: id: %d, address: %s", chainId, contract)
	}

	hasWallet, err := n.GetChainByChainIdAndContractAddress(chainId, contract)
	if err != nil {
		return
	}

	if hasWallet {
		return nil
	}

	var saveChain model.Chain
	saveChain.ChainId = chainId
	saveChain.Contract = contract
	saveChain.Decimals = decimals
	saveChain.Symbol = symbol
	saveChain.Status = 1

	if err = global.NODE_DB.Create(&saveChain).Error; err != nil {
		return
	}

	return n.updateChainList(saveChain)
}

func (n *NService) updateChainList(chain model.Chain) (err error) {
	for elementKey, element := range model.ChainList {
		if element.ChainId == chain.ChainId {
			model.ChainList[elementKey].Coins = append(model.ChainList[elementKey].Coins, model.Coin{
				Symbol:     chain.Symbol,
				Decimals:   chain.Decimals,
				Contract:   chain.Contract,
				IsMainCoin: false,
			})
			break
		}
	}

	rd, _ := json.Marshal(model.ChainList)
	global.NODE_LOG.Info("UpdateChainList: " + string(rd))

	return nil
}
