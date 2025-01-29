package wallet

import (
	"context"
	"node/global"
	"node/utils"
	"strings"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

func SendTonTransfer(chainId uint, mnemonic, pub, toAddress string, sendVal string) (hash string, err error) {

	sendValInt, err := utils.FormatToOriginalValue(sendVal, 9)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return "", err
	}

	client := liteclient.NewConnectionPool()
	configUrl := "https://ton-blockchain.github.io/testnet-global.config.json"
	err = client.AddConnectionsFromConfigUrl(context.Background(), configUrl)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	api := ton.NewAPIClient(client)
	// api = api.WithRetry() // if you want automatic retries with failover to another node

	words := strings.Split(mnemonic, " ")

	w, err := wallet.FromSeed(api, words, wallet.V3)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	// we need fresh block info to run get methods
	block, err := api.CurrentMasterchainInfo(context.Background())
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	balance, err := w.GetBalance(context.Background(), block)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	if balance.Nano().Uint64() >= sendValInt.Uint64() {
		addr := address.MustParseAddr(toAddress)
		err = w.Transfer(context.Background(), addr, tlb.MustFromTON(sendVal), "Hey bro, nice to meet you!")
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			return
		}
	}

	return "", nil
}

func SendTonTokenTransfer(chainId uint, mnemonic, pub, toAddress, coin string, sendVal string) (hash string, err error) {
	return "", nil
}
