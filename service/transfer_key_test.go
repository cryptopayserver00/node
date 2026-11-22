package service

import (
	"node/model/node/request"
	"testing"
)

func TestBuildTransferKey_IndexAndFingerprint(t *testing.T) {
	withIdx := request.NotificationRequest{
		FromAddress: "a",
		ToAddress:   "b",
		Amount:      "1",
		TransactType: "receive",
	}
	withIdx.SetTransferIndex(3)
	if got := BuildTransferKey(withIdx); got != "idx:3" {
		t.Fatalf("expected idx:3, got %s", got)
	}

	a := request.NotificationRequest{
		FromAddress:  "From",
		ToAddress:    "To",
		Amount:       "1.5",
		TransactType: "send",
	}
	b := request.NotificationRequest{
		FromAddress:  "from",
		ToAddress:    "to",
		Amount:       "1.5",
		TransactType: "SEND",
	}
	if BuildTransferKey(a) != BuildTransferKey(b) {
		t.Fatal("fingerprint should be case-insensitive for addresses/type")
	}

	c := a
	c.Amount = "1.6"
	if BuildTransferKey(a) == BuildTransferKey(c) {
		t.Fatal("different amount must produce different transfer_key")
	}
}
