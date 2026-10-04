package wallet

import (
	"context"
	"node/global/constant"
	"os"
	"testing"
)

func TestSendTonTransferLive(t *testing.T) {
	mnemonic, pub, to := os.Getenv("TON_TEST_MNEMONIC"), os.Getenv("SOL_TEST_PUB"), os.Getenv("SOL_TEST_TO")
	if mnemonic == "" || pub == "" || to == "" {
		t.Skip("set TON_TEST_MNEMONIC, SOL_TEST_PUB and SOL_TEST_TO to run against testnet")
	}
	txid, err := SendTonTransfer(context.Background(), constant.TON_TESTNET, mnemonic, pub, to, "0.1")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	t.Logf("broadcasted txid: %s", txid)
}

func TestSendTonTokenTransferLive(t *testing.T) {
	mnemonic, pub, to := os.Getenv("TON_TEST_MNEMONIC"), os.Getenv("SOL_TEST_PUB"), os.Getenv("SOL_TEST_TO")
	if mnemonic == "" || pub == "" || to == "" {
		t.Skip("set TON_TEST_MNEMONIC, SOL_TEST_PUB and SOL_TEST_TO to run against testnet")
	}
	txid, err := SendTonTokenTransfer(context.Background(), constant.TON_TESTNET, mnemonic, pub, to, constant.USDT, "1")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	t.Logf("broadcasted txid: %s", txid)
}
