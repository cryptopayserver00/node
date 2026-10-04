package utils

import (
	"crypto/rand"
	"math/bits"
	mrand "math/rand/v2"
	"strings"

	"github.com/gagliardetto/solana-go"
)

const (
	digitCharset        = "0123456789"
	alphanumericCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// randomString 使用 crypto/rand 生成指定长度的随机字符串。
// charset 必须是 ASCII 字符集，长度不超过 256。
// 采用拒绝采样，避免取模带来的分布偏差。
func randomString(length int, charset string) string {
	size := len(charset)
	if length <= 0 || size == 0 {
		return ""
	}
	if size > 256 {
		panic("utils: charset too large")
	}

	// 覆盖 [0, size) 的最小位掩码，例如 size=62 时 mask=63
	mask := byte(uint(1)<<bits.Len(uint(size-1)) - 1)

	out := make([]byte, 0, length)
	// 一次多读一些，减少被拒绝后的重复调用
	buf := make([]byte, length+length/2+8)

	for len(out) < length {
		if _, err := rand.Read(buf); err != nil {
			// Go 1.24+ 中 rand.Read 不会返回错误；
			// 更低版本出错说明系统随机源不可用，无法安全继续
			panic(err)
		}
		for _, b := range buf {
			b &= mask
			if int(b) >= size {
				continue // 拒绝，保证均匀分布
			}
			out = append(out, charset[b])
			if len(out) == length {
				break
			}
		}
	}
	return string(out)
}

// GenerateNumberRandomly 生成 prefix + length 位随机数字串。
func GenerateNumberRandomly(prefix string, length int) string {
	return prefix + randomString(length, digitCharset)
}

// GenerateStringRandomly 生成 prefix + length 位随机字母数字串。
func GenerateStringRandomly(prefix string, length int) string {
	return prefix + randomString(length, alphanumericCharset)
}

// RandomElement 从切片中随机取一个元素，切片为空时返回 false。
// 仅用于非安全场景（如随机展示、负载打散）。
func RandomElement[T any](s []T) (T, bool) {
	var zero T
	if len(s) == 0 {
		return zero, false
	}
	return s[mrand.IntN(len(s))], true
}

// GetRandomValueFromStringArray 保持原有签名，便于平滑迁移。
func GetRandomValueFromStringArray(strs []string) string {
	v, _ := RandomElement(strs)
	return v
}

func RemoveDuplicatesForString(arr []string) []string {
	encountered := map[string]bool{}

	result := []string{}

	for _, v := range arr {
		lowerCaseValue := strings.ToLower(v)
		if !encountered[lowerCaseValue] {
			result = append(result, v)
			encountered[lowerCaseValue] = true
		}
	}

	return result
}

func RemoveDuplicatesForSolanaPublicKey(addresses []solana.PublicKey) []solana.PublicKey {
	seen := make(map[string]bool)
	result := make([]solana.PublicKey, 0)

	for _, addr := range addresses {
		addrStr := addr.String()
		if !seen[addrStr] {
			seen[addrStr] = true
			result = append(result, addr)
		}
	}
	return result
}

func TruncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "...[truncated]"
}
