package wallet

import (
	"context"
	"node/global/constant"
	"os"
	"testing"
)

func TestSendTRONTransferLive(t *testing.T) {
	wif, to := os.Getenv("TRON_TEST_WIF"), os.Getenv("TRON_TEST_TO")
	if wif == "" || to == "" {
		t.Skip("set TRON_TEST_WIF and TRON_TEST_TO to run against testnet")
	}
	txid, err := SendTrxTransfer(context.Background(), constant.TRON_NILE, wif, "", to, "0.0001")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	// txid, err := SendTronTokenTransfer(context.Background(), constant.TRON_NILE, wif, "", to, "", "0.0001")
	// if err != nil {
	// t.Fatalf("send: %v", err)
	// }
	t.Logf("broadcasted txid: %s", txid)
}
