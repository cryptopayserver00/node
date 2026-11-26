package service

import (
	"math/big"
	"testing"
)

func TestValidateFreeCoinAmount(t *testing.T) {
	max := mustRat("0.01")
	tests := []struct {
		name    string
		amount  string
		max     *big.Rat
		wantErr bool
	}{
		{name: "ok equal", amount: "0.01", max: max, wantErr: false},
		{name: "ok under", amount: "0.001", max: max, wantErr: false},
		{name: "over", amount: "0.02", max: max, wantErr: true},
		{name: "zero", amount: "0", max: max, wantErr: true},
		{name: "negative", amount: "-1", max: max, wantErr: true},
		{name: "bad amount", amount: "abc", max: max, wantErr: true},
		{name: "nil max", amount: "0.01", max: nil, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFreeCoinAmount(tt.amount, tt.max)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateFreeCoinAmount(%q) err=%v wantErr=%v", tt.amount, err, tt.wantErr)
			}
		})
	}
}

func TestResolveFreeCoinMaxAmountFromUsdStable(t *testing.T) {
	max, err := resolveFreeCoinMaxAmount(t.Context(), "USDT", "100", "")
	if err != nil {
		t.Fatal(err)
	}
	if max.Cmp(mustRat("100")) != 0 {
		t.Fatalf("stable max = %s, want 100", max.FloatString(8))
	}
}

func TestResolveFreeCoinMaxAmountLegacyAbsolute(t *testing.T) {
	max, err := resolveFreeCoinMaxAmount(t.Context(), "ETH", "", "0.05")
	if err != nil {
		t.Fatal(err)
	}
	if max.Cmp(mustRat("0.05")) != 0 {
		t.Fatalf("legacy max = %s, want 0.05", max.FloatString(8))
	}
}

func TestNormalizeFreeCoinSymbol(t *testing.T) {
	if got := normalizeFreeCoinSymbol("weth"); got != "ETH" {
		t.Fatalf("got %s", got)
	}
	if got := normalizeFreeCoinSymbol("WBTC"); got != "BTC" {
		t.Fatalf("got %s", got)
	}
}

func mustRat(s string) *big.Rat {
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		panic("bad rat: " + s)
	}
	return r
}
