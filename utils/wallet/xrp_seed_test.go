package wallet

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestXrpFamilySeedMatchesFreecoinConfig(t *testing.T) {
	cfgPath := os.Getenv("NODE_CONFIG")
	candidates := []string{cfgPath, "config.yaml", "../../config.yaml"}
	var b []byte
	var err error
	for _, p := range candidates {
		if strings.TrimSpace(p) == "" {
			continue
		}
		b, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if err != nil || len(b) == 0 {
		t.Skip("config.yaml not available")
	}
	var cfg struct {
		FreeCoin struct {
			Xrp struct {
				PublicKey  string `yaml:"public_key"`
				PrivateKey string `yaml:"private_key"`
				Mnemonic   string `yaml:"mnemonic"`
			} `yaml:"xrp"`
		} `yaml:"freecoin"`
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		t.Fatal(err)
	}
	secret := strings.TrimSpace(cfg.FreeCoin.Xrp.Mnemonic)
	if secret == "" {
		secret = strings.TrimSpace(cfg.FreeCoin.Xrp.PrivateKey)
	}
	pub := strings.TrimSpace(cfg.FreeCoin.Xrp.PublicKey)
	if secret == "" || pub == "" {
		t.Skip("xrp freecoin secret/pub not configured")
	}

	priv, pk, id, err := xrpParseKey(secret, pub)
	if err != nil {
		t.Fatalf("xrpParseKey: %v", err)
	}
	if priv == nil || len(pk) == 0 {
		t.Fatal("empty key material")
	}
	addr := xrpEncodeAddress(id)
	if addr != pub {
		t.Fatalf("derived address %s != configured public_key %s", addr, pub)
	}
}
