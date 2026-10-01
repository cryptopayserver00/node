package wallet

import (
	"context"
	"node/global/constant"
	"os"
	"testing"
)

func TestSendEthTransferLive(t *testing.T) {
	pri, pub, to := os.Getenv("ETH_TEST_PRI"), os.Getenv("ETH_TEST_PUB"), os.Getenv("ETH_TEST_TO")
	if pri == "" || pub == "" || to == "" {
		t.Skip("set ETH_TEST_PRI, ETH_TEST_PUB and ETH_TEST_TO to run against testnet")
	}
	txid, err := SendEthTransfer(context.Background(), constant.ETH_SEPOLIA, pri, pub, to, "0.0001")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	t.Logf("broadcasted txid: %s", txid)
}

func TestSendEthTokenTransferLive(t *testing.T) {
	pri, pub, to := os.Getenv("ETH_TEST_PRI"), os.Getenv("ETH_TEST_PUB"), os.Getenv("ETH_TEST_TO")
	if pri == "" || pub == "" || to == "" {
		t.Skip("set ETH_TEST_PRI, ETH_TEST_PUB and ETH_TEST_TO to run against testnet")
	}
	txid, err := SendEthTokenTransfer(context.Background(), constant.ETH_SEPOLIA, pri, pub, to, constant.USDT, "1")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	t.Logf("broadcasted txid: %s", txid)
}
