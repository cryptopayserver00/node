package wallet

import (
	"context"
	"node/global/constant"
	"os"
	"testing"
)

func TestSendLtcTransferLive(t *testing.T) {
	wif, to := os.Getenv("LTC_TEST_WIF"), os.Getenv("LTC_TEST_TO")
	if wif == "" || to == "" {
		t.Skip("set LTC_TEST_WIF and LTC_TEST_TO to run against testnet")
	}
	txid, err := SendLtcTransfer(context.Background(), constant.LTC_TESTNET, wif, "", to, "0.0001")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	t.Logf("broadcasted txid: %s", txid)
}
