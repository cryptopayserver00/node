package wallet

import (
	"context"
	"node/global/constant"
	"os"
	"testing"
)

func TestSendXRPTransferLive(t *testing.T) {
	wif, to := os.Getenv("XRP_TEST_WIF"), os.Getenv("XRP_TEST_TO")
	if wif == "" || to == "" {
		t.Skip("set XRP_TEST_WIF and XRP_TEST_TO to run against testnet")
	}
	txid, err := SendXrpTransfer(context.Background(), constant.XRP_TESTNET, wif, "", to, "0.0001")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	t.Logf("broadcasted txid: %s", txid)
}
