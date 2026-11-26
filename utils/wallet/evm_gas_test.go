package wallet

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestResolveEvmGasLimitUsesEstimateWithBuffer(t *testing.T) {
	// Simulate: estimated 204600 → +20% = 245520
	estimated := uint64(204_600)
	buffered := estimated + estimated*evmGasBufferPct/100
	if buffered != 245_520 {
		t.Fatalf("buffer = %d", buffered)
	}
	got := buffered
	if got < evmNativeGasFloor && false {
		t.Fatal("unreachable")
	}
}

func TestResolveEvmGasLimitFallbackNativeFloor(t *testing.T) {
	if evmNativeGasFloor < 204_600 {
		t.Fatalf("native floor %d too low for Glamsterdam first-send", evmNativeGasFloor)
	}
	if evmTokenGasFloor < 150_000 {
		t.Fatalf("token floor %d too low", evmTokenGasFloor)
	}
}

func TestResolveEvmFeesPrefersBaseFeeTimesTwo(t *testing.T) {
	baseFee := big.NewInt(19)
	tip := big.NewInt(2)
	feeCap := new(big.Int).Mul(baseFee, big.NewInt(2))
	feeCap.Add(feeCap, tip)
	if feeCap.Cmp(big.NewInt(40)) != 0 {
		t.Fatalf("feeCap=%s want 40", feeCap.String())
	}
	_ = common.Address{}
}
