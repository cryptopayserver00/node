package wallet

import (
	"errors"
	"fmt"
	"node/global"
	"node/global/constant"
	sweepUtils "node/sweep/utils"
	"node/utils"
)

func TransferFreeCoinToReceiveAddress(chainId uint, coin, address, amount string) (hash string, err error) {
	switch chainId {
	case constant.BTC_TESTNET:
		break
	case constant.LTC_TESTNET:
		break
	case constant.ETH_SEPOLIA, constant.OP_SEPOLIA, constant.ARBITRUM_SEPOLIA:
		switch coin {
		case constant.ETH:
			hash, err = SendEthTransfer(chainId, global.NODE_CONFIG.FreeCoin.Ethereum.PrivateKey, global.NODE_CONFIG.FreeCoin.Ethereum.PublicKey, address, amount)
		default:
			hash, err = SendTokenTransfer(chainId, global.NODE_CONFIG.FreeCoin.Ethereum.PrivateKey, global.NODE_CONFIG.FreeCoin.Ethereum.PublicKey, address, coin, amount)
		}
	case constant.BSC_TESTNET:
		switch coin {
		case constant.BNB:
			hash, err = SendEthTransfer(chainId, global.NODE_CONFIG.FreeCoin.Ethereum.PrivateKey, global.NODE_CONFIG.FreeCoin.Ethereum.PublicKey, address, amount)
		default:
			hash, err = SendTokenTransfer(chainId, global.NODE_CONFIG.FreeCoin.Ethereum.PrivateKey, global.NODE_CONFIG.FreeCoin.Ethereum.PublicKey, address, coin, amount)
		}
	case constant.TRON_NILE:
		break
	}

	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	if hash == "" {
		err = errors.New("no transactions were executed")
		return
	}

	return hash, nil
}

func SendEthTransfer(chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {
	rpc := constant.GetRPCUrlByNetwork(chainId)
	if rpc == "" {
		err = errors.New("chain not support")
		return
	}

	var (
		gasLimit uint64 = 21000
		decimals int    = 18
	)

	sendValue, err := utils.FormatToOriginalValue(sendVal, decimals)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return "", err
	}

	hash, err = CallEthTransfer(chainId, rpc, pri, pub, toAddress, sendValue, gasLimit)
	return
}

func SendTokenTransfer(chainId uint, pri, pub, toAddress, coin string, sendVal string) (hash string, err error) {
	rpc := constant.GetRPCUrlByNetwork(chainId)
	if rpc == "" {
		err = errors.New("chain not support")
		return
	}

	isSupport, _, contractAddress, decimals := sweepUtils.GetContractInfoByChainIdAndSymbol(chainId, coin)
	if !isSupport {
		return "", errors.New("contract address not found")
	}

	var (
		gasLimit uint64 = 96000
	)

	sendValue, err := utils.FormatToOriginalValue(sendVal, decimals)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return "", err
	}

	hash, err = CallTokenTransfer(chainId, rpc, pri, pub, toAddress, contractAddress, sendValue, gasLimit)
	return
}
