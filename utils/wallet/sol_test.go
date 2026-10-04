package wallet

import (
	"context"
	"node/global/constant"
	"os"
	"testing"
)

func TestSendSolTransferLive(t *testing.T) {
	pri, pub, to := os.Getenv("SOL_TEST_PRI"), os.Getenv("SOL_TEST_PUB"), os.Getenv("SOL_TEST_TO")
	if pri == "" || pub == "" || to == "" {
		t.Skip("set SOL_TEST_PRI, SOL_TEST_PUB and SOL_TEST_TO to run against testnet")
	}
	txid, err := SendSolTransfer(context.Background(), constant.SOL_DEVNET, pri, pub, to, "0.001")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	t.Logf("broadcasted txid: %s", txid)
}

func TestSendSolTokenTransferLive(t *testing.T) {
	pri, pub, to := os.Getenv("SOL_TEST_PRI"), os.Getenv("SOL_TEST_PUB"), os.Getenv("SOL_TEST_TO")
	if pri == "" || pub == "" || to == "" {
		t.Skip("set SOL_TEST_PRI, SOL_TEST_PUB and SOL_TEST_TO to run against testnet")
	}
	txid, err := SendSolTokenTransfer(context.Background(), constant.SOL_DEVNET, pri, pub, to, constant.USDT, "1")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	t.Logf("broadcasted txid: %s", txid)
}
