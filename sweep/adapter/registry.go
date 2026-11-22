package adapter

import (
	"context"
	"node/global"
	"node/global/constant"
	"node/sweep/mainnet"
	"node/sweep/testnet"
)

// EnabledAdapters 根据配置与主网/测试网开关组装可运行的扫块适配器。
func EnabledAdapters() []Adapter {
	cfg := global.NODE_CONFIG.Blockchain
	mainnetMode := cfg.SweepMainnet
	var out []Adapter

	addEVM := func(enabled bool, mainID, testID uint) {
		if !enabled {
			return
		}
		id := mainID
		if !mainnetMode {
			if testID == 0 {
				return
			}
			id = testID
		}
		out = append(out, NewEVMAdapter(id))
	}

	addEVM(cfg.Ethereum, constant.ETH_MAINNET, constant.ETH_SEPOLIA)
	addEVM(cfg.Bsc, constant.BSC_MAINNET, constant.BSC_TESTNET)
	addEVM(cfg.Op, constant.OP_MAINNET, constant.OP_SEPOLIA)
	addEVM(cfg.ArbitrumOne, constant.ARBITRUM_ONE, constant.ARBITRUM_SEPOLIA)
	if cfg.ArbitrumNova && mainnetMode {
		out = append(out, NewEVMAdapter(constant.ARBITRUM_NOVA))
	}
	addEVM(cfg.Pol, constant.POL_MAINNET, constant.POL_TESTNET)
	addEVM(cfg.Avax, constant.AVAX_MAINNET, constant.AVAX_TESTNET)
	addEVM(cfg.Base, constant.BASE_MAINNET, constant.BASE_SEPOLIA)

	if cfg.Bitcoin {
		out = append(out, NewFuncAdapter(pickChain(mainnetMode, constant.BTC_MAINNET, constant.BTC_TESTNET), "bitcoin", func(ctx context.Context) {
			if mainnetMode {
				mainnet.SweepBtcBlockchain(ctx)
			} else {
				testnet.SweepBtcTestnetBlockchain(ctx)
			}
		}))
	}
	if cfg.Litecoin {
		out = append(out, NewFuncAdapter(pickChain(mainnetMode, constant.LTC_MAINNET, constant.LTC_TESTNET), "litecoin", func(ctx context.Context) {
			if mainnetMode {
				mainnet.SweepLtcBlockchain(ctx)
			} else {
				testnet.SweepLtcTestnetBlockchain(ctx)
			}
		}))
	}
	if cfg.Bch && mainnetMode {
		out = append(out, NewFuncAdapter(constant.BCH_MAINNET, "bitcoincash", mainnet.SweepBchBlockchain))
	}
	if cfg.Tron {
		out = append(out, NewFuncAdapter(pickChain(mainnetMode, constant.TRON_MAINNET, constant.TRON_NILE), "tron", func(ctx context.Context) {
			if mainnetMode {
				mainnet.SweepTronBlockchain(ctx)
			} else {
				testnet.SweepTronNileBlockchain(ctx)
			}
		}))
	}
	if cfg.Solana {
		out = append(out, NewFuncAdapter(pickChain(mainnetMode, constant.SOL_MAINNET, constant.SOL_DEVNET), "solana", func(ctx context.Context) {
			if mainnetMode {
				mainnet.SweepSolBlockchain(ctx)
			} else {
				testnet.SweepSolDevnetBlockchain(ctx)
			}
		}))
	}
	if cfg.Ton {
		out = append(out, NewFuncAdapter(pickChain(mainnetMode, constant.TON_MAINNET, constant.TON_TESTNET), "ton", func(ctx context.Context) {
			if mainnetMode {
				mainnet.SweepTonBlockchain(ctx)
			} else {
				testnet.SweepTonTestnetBlockchain(ctx)
			}
		}))
	}
	if cfg.Xrp {
		out = append(out, NewFuncAdapter(pickChain(mainnetMode, constant.XRP_MAINNET, constant.XRP_TESTNET), "xrp", func(ctx context.Context) {
			if mainnetMode {
				mainnet.SweepXrpBlockchain(ctx)
			} else {
				testnet.SweepXrpTestnetBlockchain(ctx)
			}
		}))
	}

	return out
}

func pickChain(mainnetMode bool, mainID, testID uint) uint {
	if mainnetMode {
		return mainID
	}
	return testID
}
