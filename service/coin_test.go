package service

import "testing"

func TestValidateFreeCoinAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  string
		max     string
		wantErr bool
	}{
		{name: "ok equal", amount: "0.01", max: "0.01", wantErr: false},
		{name: "ok under", amount: "0.001", max: "0.01", wantErr: false},
		{name: "over", amount: "0.02", max: "0.01", wantErr: true},
		{name: "zero", amount: "0", max: "0.01", wantErr: true},
		{name: "negative", amount: "-1", max: "0.01", wantErr: true},
		{name: "bad amount", amount: "abc", max: "0.01", wantErr: true},
		{name: "bad max", amount: "0.01", max: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFreeCoinAmount(tt.amount, tt.max)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateFreeCoinAmount(%q, %q) err=%v wantErr=%v", tt.amount, tt.max, err, tt.wantErr)
			}
		})
	}
}
