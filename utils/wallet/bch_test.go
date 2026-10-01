package wallet

import (
	"context"
	"node/global/constant"
	"os"
	"testing"
)

func TestSendBchTransferLive(t *testing.T) {
	wif, to := os.Getenv("BCH_TEST_WIF"), os.Getenv("BCH_TEST_TO")
	if wif == "" || to == "" {
		t.Skip("set BCH_TEST_WIF and BCH_TEST_TO to run against testnet")
	}
	txid, err := SendBchTransfer(context.Background(), constant.BCH_TESTNET, wif, "", to, "0.0001")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	t.Logf("broadcasted txid: %s", txid)
}
