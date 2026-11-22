package model

import (
	"encoding/json"
	"testing"
	"time"

	"node/model/common"
)

func TestOwnTransactionMarshalJSONIncludesID(t *testing.T) {
	tx := OwnTransaction{
		NODE_MODEL: common.NODE_MODEL{
			ID:        42,
			CreatedAt: time.Unix(1700000000, 0).UTC(),
		},
		ChainId: 1,
		Hash:    "0xabc",
		Address: "0xaddr",
	}
	b, err := json.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["id"].(float64) != 42 {
		t.Fatalf("expected id=42, got %v", m["id"])
	}
	if m["hash"] != "0xabc" {
		t.Fatalf("expected hash, got %v", m["hash"])
	}
}
