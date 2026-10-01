package wallet

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// 金额：十进制字符串 -> 最小单位，避免 float 精度问题
// ---------------------------------------------------------------------------

// parseDecimal 把 "1.5" 按 decimals 位小数换算成最小单位。金额必须为正数。
func parseDecimal(s string, decimals int) (*big.Int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("empty amount")
	}
	whole, frac, _ := strings.Cut(s, ".")
	if whole == "" {
		whole = "0"
	}
	if len(frac) > decimals {
		return nil, fmt.Errorf("amount has more than %d decimals: %s", decimals, s)
	}
	frac += strings.Repeat("0", decimals-len(frac))
	if !isDigits(whole) || (frac != "" && !isDigits(frac)) {
		return nil, fmt.Errorf("invalid amount: %s", s)
	}
	v, ok := new(big.Int).SetString(whole+frac, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount: %s", s)
	}
	if v.Sign() <= 0 {
		return nil, errors.New("amount must be positive")
	}
	return v, nil
}

// ---------------------------------------------------------------------------
// 时钟（测试里可替换）
// ---------------------------------------------------------------------------

var nowFn = time.Now

// ---------------------------------------------------------------------------
// 公钥比对：配置里的 PublicKey 用来确认私钥/助记词没有填错
// ---------------------------------------------------------------------------

// matchPubKeyHex 判断 pub（hex，可带 0x）是否等于 candidates 中任意一个。
// 空字符串视为不校验，返回 true。
func matchPubKeyHex(pub string, candidates ...[]byte) bool {
	pub = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(pub), "0x"))
	if pub == "" {
		return true
	}
	for _, c := range candidates {
		if pub == hex.EncodeToString(c) {
			return true
		}
	}
	return false
}
