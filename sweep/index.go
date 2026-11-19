package sweep

import (
	"context"
	"node/global"
	"node/global/constant"
	"node/sweep/core"
	"node/sweep/mainnet"
	"node/sweep/setup"
	"node/sweep/testnet"
)

func RunBlockSweep(ctx context.Context) {
	setup.SetupPublicKey(ctx)

	if global.NODE_CONFIG.Blockchain.Ethereum {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			core.RunEVMSweep(ctx, constant.ETH_MAINNET)
		} else {
			core.RunEVMSweep(ctx, constant.ETH_SEPOLIA)
		}
	}

	if global.NODE_CONFIG.Blockchain.Bsc {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			core.RunEVMSweep(ctx, constant.BSC_MAINNET)
		} else {
			core.RunEVMSweep(ctx, constant.BSC_TESTNET)
		}
	}

	if global.NODE_CONFIG.Blockchain.Bitcoin {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepBtcBlockchain(ctx)
		} else {
			testnet.SweepBtcTestnetBlockchain(ctx)
		}
	}

	if global.NODE_CONFIG.Blockchain.Tron {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepTronBlockchain(ctx)
		} else {
			testnet.SweepTronNileBlockchain(ctx)
		}
	}

	if global.NODE_CONFIG.Blockchain.Litecoin {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepLtcBlockchain(ctx)
		} else {
			testnet.SweepLtcTestnetBlockchain(ctx)
		}
	}

	if global.NODE_CONFIG.Blockchain.Op {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			core.RunEVMSweep(ctx, constant.OP_MAINNET)
		} else {
			core.RunEVMSweep(ctx, constant.OP_SEPOLIA)
		}
	}

	if global.NODE_CONFIG.Blockchain.ArbitrumOne {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			core.RunEVMSweep(ctx, constant.ARBITRUM_ONE)
		} else {
			core.RunEVMSweep(ctx, constant.ARBITRUM_SEPOLIA)
		}
	}

	if global.NODE_CONFIG.Blockchain.ArbitrumNova {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			core.RunEVMSweep(ctx, constant.ARBITRUM_NOVA)
		}
	}

	if global.NODE_CONFIG.Blockchain.Solana {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepSolBlockchain(ctx)
		} else {
			testnet.SweepSolDevnetBlockchain(ctx)
		}
	}

	if global.NODE_CONFIG.Blockchain.Ton {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepTonBlockchain(ctx)
		} else {
			testnet.SweepTonTestnetBlockchain(ctx)
		}
	}

	if global.NODE_CONFIG.Blockchain.Xrp {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepXrpBlockchain(ctx)
		} else {
			testnet.SweepXrpTestnetBlockchain(ctx)
		}
	}

	if global.NODE_CONFIG.Blockchain.Bch {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepBchBlockchain(ctx)
		}
	}

	if global.NODE_CONFIG.Blockchain.Pol {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			core.RunEVMSweep(ctx, constant.POL_MAINNET)
		} else {
			core.RunEVMSweep(ctx, constant.POL_TESTNET)
		}
	}

	if global.NODE_CONFIG.Blockchain.Avax {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			core.RunEVMSweep(ctx, constant.AVAX_MAINNET)
		} else {
			core.RunEVMSweep(ctx, constant.AVAX_TESTNET)
		}
	}

	if global.NODE_CONFIG.Blockchain.Base {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			core.RunEVMSweep(ctx, constant.BASE_MAINNET)
		} else {
			core.RunEVMSweep(ctx, constant.BASE_SEPOLIA)
		}
	}
}
