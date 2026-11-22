package outbox

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestSignBody(t *testing.T) {
	secret := "test-secret"
	ts := "1700000000"
	body := []byte(`{"id":1}`)
	got := signBody(secret, ts, body)

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(ts))
	_, _ = mac.Write([]byte{'.'})
	_, _ = mac.Write(body)
	want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if got != want {
		t.Fatalf("sign mismatch: got %s want %s", got, want)
	}
}

func TestVerifySignature(t *testing.T) {
	secret := "test-secret"
	ts := "1700000000"
	body := []byte(`{"id":1}`)
	sig := signBody(secret, ts, body)
	if !VerifySignature(secret, ts, sig, body) {
		t.Fatal("expected valid signature")
	}
	if VerifySignature(secret, ts, sig, []byte(`{"id":2}`)) {
		t.Fatal("tampered body must fail")
	}
	if VerifySignature(secret, "1", sig, body) {
		t.Fatal("tampered timestamp must fail")
	}
}
