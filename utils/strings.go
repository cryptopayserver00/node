package utils

import (
	"math/rand"
	"strings"
	"time"
)

var charset = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func GenerateStringRandomly(prefix string, length int) string {
	return prefix + StringWithCharset(length, charset)
}

func StringWithCharset(length int, charset []rune) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
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
