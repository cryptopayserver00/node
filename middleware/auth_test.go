package middleware

import "testing"

func TestSecureEqual(t *testing.T) {
	if secureEqual("", "a") {
		t.Fatal("empty should not match")
	}
	if secureEqual("a", "") {
		t.Fatal("empty expected should not match")
	}
	if !secureEqual("secret", "secret") {
		t.Fatal("equal tokens should match")
	}
	if secureEqual("secret", "Secret") {
		t.Fatal("different tokens should not match")
	}
}
