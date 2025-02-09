package wallet

import (
	"context"
	"encoding/base64"
	"errors"
	"math/rand/v2"
	"node/global"
	"node/global/constant"
	sweepUtils "node/sweep/utils"
	"node/utils"
	"strings"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func SendTonTransfer(chainId uint, mnemonic, pub, toAddress string, sendVal string) (hash string, err error) {

	sendValInt, err := utils.FormatToOriginalValue(sendVal, 9)
	if err != nil {
		return
	}

	client := liteclient.NewConnectionPool()

	url := constant.GetHttpUrlByNetwork(chainId)
	if url == "" {
		err = errors.New("chain not support")
		return
	}

	cfg, err := liteclient.GetConfigFromUrl(context.Background(), url)
	if err != nil {
		return
	}

	err = client.AddConnectionsFromConfig(context.Background(), cfg)
	if err != nil {
		return
	}

	api := ton.NewAPIClient(client, ton.ProofCheckPolicyFast).WithRetry()
	api.SetTrustedBlockFromConfig(cfg)

	ctx := client.StickyContext(context.Background())

	words := strings.Split(mnemonic, " ")

	w, err := wallet.FromSeed(api, words, wallet.ConfigV5R1Final{
		NetworkGlobalID: wallet.TestnetGlobalID,
	})

	if err != nil {
		return
	}

	err = client.AddConnectionsFromConfigUrl(context.Background(), url)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	block, err := api.CurrentMasterchainInfo(context.Background())
	if err != nil {
		return
	}

	balance, err := w.GetBalance(ctx, block)
	if err != nil {
		return
	}

	if balance.Nano().Uint64() >= sendValInt.Uint64() {
		addr := address.MustParseAddr(toAddress)

		// if destination wallet is not initialized (or you don't care)
		// you should set bounce to false to not get money back.
		// If bounce is true, money will be returned in case of not initialized destination wallet or smart-contract error
		bounce := false

		transfer, innerErr := w.BuildTransfer(addr, tlb.MustFromTON(sendVal), bounce, "Hey bro, nice to meet you!")

		if innerErr != nil {
			return
		}

		tx, _, innerErr := w.SendWaitTransaction(ctx, transfer)
		if innerErr != nil {
			return
		}

		hash = base64.StdEncoding.EncodeToString(tx.Hash)

	}

	return hash, nil
}

func SendTonTokenTransfer(chainId uint, mnemonic, pub, toAddress, coin string, sendVal string) (hash string, err error) {

	isSupport, _, contractAddress, decimals := sweepUtils.GetContractInfoByChainIdAndSymbol(chainId, coin)
	if !isSupport {
		return "", errors.New("contract address not found")
	}

	sendValInt, err := utils.FormatToOriginalValue(sendVal, decimals)
	if err != nil {
		return
	}

	client := liteclient.NewConnectionPool()

	url := constant.GetHttpUrlByNetwork(chainId)
	if url == "" {
		err = errors.New("chain not support")
		return
	}

	cfg, err := liteclient.GetConfigFromUrl(context.Background(), url)
	if err != nil {
		return
	}

	err = client.AddConnectionsFromConfig(context.Background(), cfg)
	if err != nil {
		return
	}

	api := ton.NewAPIClient(client, ton.ProofCheckPolicyFast).WithRetry()
	api.SetTrustedBlockFromConfig(cfg)

	ctx := client.StickyContext(context.Background())

	words := strings.Split(mnemonic, " ")

	w, err := wallet.FromSeed(api, words, wallet.ConfigV5R1Final{
		NetworkGlobalID: wallet.TestnetGlobalID,
	})

	if err != nil {
		return
	}

	err = client.AddConnectionsFromConfigUrl(context.Background(), url)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	block, err := api.CurrentMasterchainInfo(context.Background())
	if err != nil {
		return
	}

	balance, err := w.GetBalance(ctx, block)
	if err != nil {
		return
	}

	if balance.Nano().Uint64() >= sendValInt.Uint64() {
		// create transaction body cell, depends on what contract needs, just random example here
		body := cell.BeginCell().
			MustStoreUInt(0x123abc55, 32).    // op code
			MustStoreUInt(rand.Uint64(), 64). // query id
			// payload:
			MustStoreAddr(address.MustParseAddr(contractAddress)).
			MustStoreRef(
				cell.BeginCell().
					MustStoreBigCoins(tlb.MustFromTON("1.521").Nano()).
					EndCell(),
			).EndCell()

		/*
			// alternative, more high level way to serialize cell; see tlb.LoadFromCell method for doc
			type ContractRequest struct {
				_        tlb.Magic        `tlb:"#123abc55"`
				QueryID  uint64           `tlb:"## 64"`
				Addr     *address.Address `tlb:"addr"`
				RefMoney tlb.Coins        `tlb:"^"`
			}

			body, err := tlb.ToCell(ContractRequest{
				QueryID:  rand.Uint64(),
				Addr:     address.MustParseAddr(contractAddress),
				RefMoney: tlb.MustFromTON("1.521"),
			})
		*/

		tx, _, innerErr := w.SendWaitTransaction(context.Background(), &wallet.Message{
			Mode: wallet.PayGasSeparately, // pay fees separately (from balance, not from amount)
			InternalMessage: &tlb.InternalMessage{
				Bounce:  true, // return amount in case of processing error
				DstAddr: address.MustParseAddr(toAddress),
				Amount:  tlb.MustFromTON("0.03"),
				Body:    body,
			},
		})
		if innerErr != nil {
			return
		}

		hash = base64.StdEncoding.EncodeToString(tx.Hash)
	}

	return hash, nil
}
