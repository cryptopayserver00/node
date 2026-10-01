package wallet

import (
	"context"
	"encoding/hex"
	"node/global/constant"
	"os"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

// ---------------------------------------------------------------------------
// 测试辅助
// ---------------------------------------------------------------------------

type testWallet struct {
	wif      string
	pubHex   string // 压缩公钥 hex
	pubBytes []byte
	addr     string
	pkScript []byte
}

func newTestWallet(t *testing.T, params *chaincfg.Params) testWallet {
	t.Helper()
	priv, err := btcec.NewPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	wif, err := btcutil.NewWIF(priv, params, true)
	if err != nil {
		t.Fatal(err)
	}
	pub := priv.PubKey().SerializeCompressed()
	addr, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(pub), params)
	if err != nil {
		t.Fatal(err)
	}
	pk, err := txscript.PayToAddrScript(addr)
	if err != nil {
		t.Fatal(err)
	}
	return testWallet{
		wif:      wif.String(),
		pubHex:   hex.EncodeToString(pub),
		pubBytes: pub,
		addr:     addr.EncodeAddress(),
		pkScript: pk,
	}
}

// ---------------------------------------------------------------------------
// ParseBtcToSatoshi
// ---------------------------------------------------------------------------

func TestParseBtcToSatoshi(t *testing.T) {
	tests := []struct {
		in      string
		want    int64
		wantErr bool
	}{
		{"1", 100_000_000, false},
		{"0.001", 100_000, false},
		{"0.00000001", 1, false},
		{".5", 50_000_000, false},
		{" 0.5 ", 50_000_000, false},
		{"0.29", 29_000_000, false}, // float64: 0.29*1e8 = 28999999.999...，这里必须精确
		{"0.1", 10_000_000, false},
		{"21000000", 2_100_000_000_000_000, false},

		{"", 0, true},
		{"abc", 0, true},
		{"-1", 0, true},
		{"1e-3", 0, true},
		{"1.2.3", 0, true},
		{"0.123456789", 0, true}, // 超过 8 位小数
		{"21000001", 0, true},    // 超过总量
		{"99999999999999999999", 0, true},
	}
	for _, tc := range tests {
		got, err := ParseBtcToSatoshi(tc.in)
		if (err != nil) != tc.wantErr {
			t.Errorf("ParseBtcToSatoshi(%q) err=%v, wantErr=%v", tc.in, err, tc.wantErr)
			continue
		}
		if !tc.wantErr && got != tc.want {
			t.Errorf("ParseBtcToSatoshi(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// 可选：真实测试网集成测试（默认跳过）
//
//   BTC_TEST_WIF=<测试网压缩 WIF> BTC_TEST_TO=<tb1q... 收款地址> \
//   go test -run TestSendBtcTransferLive -v
//
// 会真实广播一笔 0.00001 tBTC 的测试网交易。
// ---------------------------------------------------------------------------

func TestSendBtcTransferLive(t *testing.T) {
	wif, to := os.Getenv("BTC_TEST_WIF"), os.Getenv("BTC_TEST_TO")
	if wif == "" || to == "" {
		t.Skip("set BTC_TEST_WIF and BTC_TEST_TO to run against testnet")
	}
	txid, err := SendBtcTransfer(context.Background(), constant.BTC_TESTNET, wif, "", to, "0.00001")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	t.Logf("broadcasted txid: %s", txid)
}
