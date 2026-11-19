package utils

import "strings"

// BuildHexAddressSet 将监控地址规范为 hex 后放入 map，供 O(1) 匹配。
func BuildHexAddressSet(publicKeys []string) map[string]struct{} {
	set := make(map[string]struct{}, len(publicKeys))
	for _, k := range publicKeys {
		set[HexToAddress(k)] = struct{}{}
	}
	return set
}

// MatchHexAddressSet 任一 candidate 落在监控集合中则返回 true。
func MatchHexAddressSet(set map[string]struct{}, candidates ...string) bool {
	for _, c := range candidates {
		if _, ok := set[HexToAddress(c)]; ok {
			return true
		}
	}
	return false
}

// BuildFoldAddressSet 大小写不敏感地址集合（BTC/Tron/Sol 等非 EVM）。
func BuildFoldAddressSet(publicKeys []string) map[string]struct{} {
	set := make(map[string]struct{}, len(publicKeys))
	for _, k := range publicKeys {
		set[strings.ToLower(strings.TrimSpace(k))] = struct{}{}
	}
	return set
}

// MatchFoldAddressSet 大小写不敏感匹配。
func MatchFoldAddressSet(set map[string]struct{}, candidates ...string) bool {
	for _, c := range candidates {
		if _, ok := set[strings.ToLower(strings.TrimSpace(c))]; ok {
			return true
		}
	}
	return false
}
