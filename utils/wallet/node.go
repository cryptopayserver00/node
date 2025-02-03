package wallet

import (
	"errors"
	"node/global"
	"node/global/constant"
)

func TransferFreeCoinToReceiveAddress(chainId uint, coin, address, amount string) (hash string, err error) {
	switch chainId {
	case constant.BTC_TESTNET:
		hash, err = SendBtcTransferByPsbt(chainId, global.NODE_CONFIG.FreeCoin.Bitcoin.PrivateKey, global.NODE_CONFIG.FreeCoin.Bitcoin.PublicKey, address, amount)
	case constant.LTC_TESTNET:
		hash, err = SendLtcTransfer(chainId, global.NODE_CONFIG.FreeCoin.Litecoin.PrivateKey, global.NODE_CONFIG.FreeCoin.Litecoin.PublicKey, address, amount)
	case constant.ETH_SEPOLIA, constant.OP_SEPOLIA, constant.ARBITRUM_SEPOLIA, constant.BASE_SEPOLIA:
		switch coin {
		case constant.ETH:
			hash, err = SendEthTransfer(chainId, global.NODE_CONFIG.FreeCoin.Ethereum.PrivateKey, global.NODE_CONFIG.FreeCoin.Ethereum.PublicKey, address, amount)
		default:
			hash, err = SendEthTokenTransfer(chainId, global.NODE_CONFIG.FreeCoin.Ethereum.PrivateKey, global.NODE_CONFIG.FreeCoin.Ethereum.PublicKey, address, coin, amount)
		}
	case constant.BSC_TESTNET:
		switch coin {
		case constant.BNB:
			hash, err = SendEthTransfer(chainId, global.NODE_CONFIG.FreeCoin.Ethereum.PrivateKey, global.NODE_CONFIG.FreeCoin.Ethereum.PublicKey, address, amount)
		default:
			hash, err = SendEthTokenTransfer(chainId, global.NODE_CONFIG.FreeCoin.Ethereum.PrivateKey, global.NODE_CONFIG.FreeCoin.Ethereum.PublicKey, address, coin, amount)
		}
	case constant.TRON_NILE:
		switch coin {
		case constant.TRX:
			hash, err = SendTrxTransfer(chainId, global.NODE_CONFIG.FreeCoin.Tron.PrivateKey, global.NODE_CONFIG.FreeCoin.Tron.PublicKey, address, amount)
		default:
			hash, err = SendTronTokenTransfer(chainId, global.NODE_CONFIG.FreeCoin.Tron.PrivateKey, global.NODE_CONFIG.FreeCoin.Tron.PublicKey, address, coin, amount)
		}
	case constant.SOL_DEVNET:
		switch coin {
		case constant.SOL:
			hash, err = SendSolTransfer(chainId, global.NODE_CONFIG.FreeCoin.Solana.PrivateKey, global.NODE_CONFIG.FreeCoin.Solana.PublicKey, address, amount)
		default:
			hash, err = SendSolTokenTransfer(chainId, global.NODE_CONFIG.FreeCoin.Solana.PrivateKey, global.NODE_CONFIG.FreeCoin.Solana.PublicKey, address, coin, amount)
		}
	case constant.TON_TESTNET:
		switch coin {
		case constant.TON:
			hash, err = SendTonTransfer(chainId, global.NODE_CONFIG.FreeCoin.Ton.Mnemonic, global.NODE_CONFIG.FreeCoin.Ton.PublicKey, address, amount)
		default:
			hash, err = SendTonTokenTransfer(chainId, global.NODE_CONFIG.FreeCoin.Ton.Mnemonic, global.NODE_CONFIG.FreeCoin.Ton.PublicKey, address, coin, amount)
		}
	case constant.XRP_TESTNET:
		hash, err = SendXrpTransfer(chainId, global.NODE_CONFIG.FreeCoin.Xrp.Mnemonic, global.NODE_CONFIG.FreeCoin.Xrp.PublicKey, address, amount)
	case constant.BCH_TESTNET:
		hash, err = SendBchTransfer(chainId, global.NODE_CONFIG.FreeCoin.Xrp.Mnemonic, global.NODE_CONFIG.FreeCoin.Xrp.PublicKey, address, amount)
	case constant.POL_TESTNET:
		switch coin {
		case constant.POL:
			hash, err = SendPolTransfer(chainId, global.NODE_CONFIG.FreeCoin.Ethereum.PrivateKey, global.NODE_CONFIG.FreeCoin.Ethereum.PublicKey, address, amount)
		default:
			hash, err = SendPolTokenTransfer(chainId, global.NODE_CONFIG.FreeCoin.Ethereum.PrivateKey, global.NODE_CONFIG.FreeCoin.Ethereum.PublicKey, address, coin, amount)
		}
	case constant.AVAX_TESTNET:
		switch coin {
		case constant.AVAX:
			hash, err = SendAvaxTransfer(chainId, global.NODE_CONFIG.FreeCoin.Ethereum.PrivateKey, global.NODE_CONFIG.FreeCoin.Ethereum.PublicKey, address, amount)
		default:
			hash, err = SendAvaxTokenTransfer(chainId, global.NODE_CONFIG.FreeCoin.Ethereum.PrivateKey, global.NODE_CONFIG.FreeCoin.Ethereum.PublicKey, address, coin, amount)
		}
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
