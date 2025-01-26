package wallet

import (
	"errors"
	"node/global"
	"node/global/constant"
)

func TransferFreeCoinToReceiveAddress(chainId uint, coin, address, amount string) (hash string, err error) {
	switch chainId {
	case constant.BTC_TESTNET:
		hash, err = SendBtcTransfer(chainId, global.NODE_CONFIG.FreeCoin.Bitcoin.PrivateKey, global.NODE_CONFIG.FreeCoin.Bitcoin.PublicKey, address, amount)
	case constant.LTC_TESTNET:
		hash, err = SendLtcTransfer(chainId, global.NODE_CONFIG.FreeCoin.Litecoin.PrivateKey, global.NODE_CONFIG.FreeCoin.Litecoin.PublicKey, address, amount)
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
	case constant.SOL_DEVNET:
		break
	case constant.TON_TESTNET:
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
