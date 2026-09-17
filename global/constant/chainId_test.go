package constant

import "testing"

func TestIsAddressSupport(t *testing.T) {
	tests := []struct {
		name    string
		chainId uint
		address string
		want    bool
	}{
		// ---------- EVM 系列（共用 common.IsHexAddress 校验规则）----------
		{"ETH mainnet 合法地址", ETH_MAINNET, "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"[:42], true},
		{"ETH mainnet 缺少0x前缀", ETH_MAINNET, "742d35Cc6634C0532925a3b844Bc454e4438f44e", false},
		{"ETH mainnet 长度不足", ETH_MAINNET, "0x1234", false},
		{"ETH mainnet 含非法字符", ETH_MAINNET, "0xZZZd35Cc6634C0532925a3b844Bc454e4438f44e", false},
		{"ETH sepolia 合法地址", ETH_SEPOLIA, "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed", true},
		{"BSC mainnet 合法地址", BSC_MAINNET, "0x178fa73B31673268D65707f14D39153f1fE38385", true},
		{"BSC testnet 空字符串", BSC_TESTNET, "", false},
		{"Arbitrum One 合法地址", ARBITRUM_ONE, "0x000000000000000000000000000000000000dEaD", true},
		{"Arbitrum Nova 合法地址", ARBITRUM_NOVA, "0x000000000000000000000000000000000000dEaD", true},
		{"Arbitrum Sepolia 合法地址", ARBITRUM_SEPOLIA, "0x000000000000000000000000000000000000dEaD", true},
		{"OP mainnet 合法地址", OP_MAINNET, "0x000000000000000000000000000000000000dEaD", true},
		{"OP sepolia 合法地址", OP_SEPOLIA, "0x000000000000000000000000000000000000dEaD", true},
		{"Polygon mainnet 合法地址", POL_MAINNET, "0x000000000000000000000000000000000000dEaD", true},
		{"Polygon testnet 合法地址", POL_TESTNET, "0x000000000000000000000000000000000000dEaD", true},
		{"Avalanche mainnet 合法地址", AVAX_MAINNET, "0x000000000000000000000000000000000000dEaD", true},
		{"Avalanche testnet 合法地址", AVAX_TESTNET, "0x000000000000000000000000000000000000dEaD", true},
		{"Base mainnet 合法地址", BASE_MAINNET, "0x000000000000000000000000000000000000dEaD", true},
		{"Base sepolia 合法地址", BASE_SEPOLIA, "0x000000000000000000000000000000000000dEaD", true},

		// ---------- BTC ----------
		{"BTC mainnet 合法 Legacy 地址", BTC_MAINNET, "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", true},
		{"BTC mainnet 合法 Bech32 地址", BTC_MAINNET, "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", true},
		{"BTC mainnet 校验和错误", BTC_MAINNET, "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN3", false},
		{"BTC mainnet 用测试网地址", BTC_MAINNET, "mipcBbFg9gMiCh81Kj8tqqdgoZub1ZJRfn", false},
		{"BTC testnet 合法地址", BTC_TESTNET, "mipcBbFg9gMiCh81Kj8tqqdgoZub1ZJRfn", true},
		{"BTC testnet 用主网地址", BTC_TESTNET, "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", false},

		// ---------- LTC ----------
		{"LTC mainnet 合法地址", LTC_MAINNET, "LMZMqqBevMzxYFqtGpudbzyqk3hfF2YKga", true},
		{"LTC mainnet 用测试网地址", LTC_MAINNET, "mmooHPyznc35Uk1UYGP4boP6Qj7iJLtsPw", false},
		{"LTC testnet 合法地址", LTC_TESTNET, "mmooHPyznc35Uk1UYGP4boP6Qj7iJLtsPw", true},
		{"LTC 非法格式（含0/O/I/l）", LTC_MAINNET, "L0OIl234567890123456789012345678", false},

		// ---------- TRON ----------
		{"TRON mainnet 合法地址", TRON_MAINNET, "TXYZopYRdj2D9XRtbG411XZZ3kM5VkAeBf", true},
		{"TRON nile 合法地址", TRON_NILE, "TXYZopYRdj2D9XRtbG411XZZ3kM5VkAeBf", true},
		{"TRON 缺少T前缀", TRON_MAINNET, "1XYZopYRdj2D9XRtbG411XZZ3kM5VkAeBf", false},
		{"TRON 空字符串", TRON_MAINNET, "", false},

		// ---------- SOL ----------
		{"SOL mainnet 合法地址（WSOL mint）", SOL_MAINNET, "So11111111111111111111111111111111111111112", true},
		{"SOL devnet 合法地址", SOL_DEVNET, "So11111111111111111111111111111111111111112", true},
		{"SOL 非base58字符", SOL_MAINNET, "So1111111111111111111111111111111111111111O", false}, // 含大写O，base58不允许
		{"SOL 长度不足", SOL_MAINNET, "abcd", false},

		// ---------- TON ----------
		{"TON mainnet 合法地址", TON_MAINNET, "EQAvDfWFG0oYX19jwNDNBBL1rKNT9XfaGP9HyTb5nb2Eml6y", true},
		{"TON testnet 合法地址", TON_TESTNET, "EQAvDfWFG0oYX19jwNDNBBL1rKNT9XfaGP9HyTb5nb2Eml6y", true},
		{"TON 非法格式", TON_MAINNET, "not-a-ton-address", false},

		// ---------- XRP ----------
		{"XRP mainnet 合法经典地址", XRP_MAINNET, "rrncdF4TSFPdk6QbDzqqDjxZgtStk1iwi7", true},
		{"XRP testnet 合法经典地址（格式通用）", XRP_TESTNET, "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", true},
		{"XRP 缺少r前缀", XRP_MAINNET, "Hb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", false},
		{"XRP 校验和错误", XRP_MAINNET, "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTi", false},
		{"XRP 空字符串", XRP_MAINNET, "", false},

		// ---------- BCH ----------
		{"BCH mainnet 合法地址", BCH_MAINNET, "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", true},
		{"BCH testnet 用主网地址", BCH_TESTNET, "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", false},

		// ---------- 边界 / 异常场景 ----------
		{"不支持的链ID", 999999, "0x0000000000000000000000000000000000dEaD", false},
		{"合法链ID但地址为空", ETH_MAINNET, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAddressSupport(tt.chainId, tt.address)
			if got != tt.want {
				t.Errorf("IsAddressSupport(chainId=%d, address=%q) = %v, want %v",
					tt.chainId, tt.address, got, tt.want)
			}
		})
	}
}

// TestLtcValidateAddress 单独测试 LTC 地址校验函数（区分主网/测试网）
func TestLtcValidateAddress(t *testing.T) {
	tests := []struct {
		name    string
		chainId uint
		address string
		want    bool
	}{
		{"主网合法Legacy地址", LTC_MAINNET, "LMZMqqBevMzxYFqtGpudbzyqk3hfF2YKga", true},
		{"主网地址传testnet参数应失败", LTC_TESTNET, "LMZMqqBevMzxYFqtGpudbzyqk3hfF2YKga", false},
		{"测试网合法地址", LTC_TESTNET, "mmooHPyznc35Uk1UYGP4boP6Qj7iJLtsPw", true},
		{"测试网地址传mainnet参数应失败", LTC_MAINNET, "mmooHPyznc35Uk1UYGP4boP6Qj7iJLtsPw", false},
		{"空字符串", LTC_MAINNET, "", false},
		{"随机乱码", LTC_MAINNET, "not-a-valid-address-at-all", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LtcValidateAddress(tt.chainId, tt.address)
			if got != tt.want {
				t.Errorf("LtcValidateAddress(chainId=%d, address=%q) = %v, want %v",
					tt.chainId, tt.address, got, tt.want)
			}
		})
	}
}

// BenchmarkIsAddressSupport 简单基准测试，关注高频调用场景下的性能（例如批量地址校验）
func BenchmarkIsAddressSupport(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsAddressSupport(ETH_MAINNET, "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"[:42])
	}
}
