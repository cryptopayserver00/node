package utils

import (
	"math/rand/v2"
	"strings"

	"github.com/gagliardetto/solana-go"
)

var charset = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func GenerateStringRandomly(prefix string, length int) string {
	return prefix + StringWithCharset(length, charset)
}

func StringWithCharset(length int, charset []rune) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))]
	}
	return string(b)
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
