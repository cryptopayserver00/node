package utils

import "testing"

func TestMatchHexAddressSet(t *testing.T) {
	set := BuildHexAddressSet([]string{
		"0xAbCDEF0000000000000000000000000000000001",
		"0x0000000000000000000000000000000000000002",
	})
	if !MatchHexAddressSet(set, "0xabcdef0000000000000000000000000000000001") {
		t.Fatal("expected match ignoring case via HexToAddress")
	}
	if MatchHexAddressSet(set, "0x0000000000000000000000000000000000000099") {
		t.Fatal("unexpected match")
	}
}

func TestMatchFoldAddressSet(t *testing.T) {
	set := BuildFoldAddressSet([]string{"Abc", "Def"})
	if !MatchFoldAddressSet(set, "abc") {
		t.Fatal("expected fold match")
	}
}
