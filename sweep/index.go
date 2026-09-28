package sweep

import (
	"context"
	"node/global"
	"node/sweep/mainnet"
	"node/sweep/setup"
	"node/sweep/testnet"
)

func RunBlockSweep(ctx context.Context) {
	setup.SetupPublicKey(ctx)

	if global.NODE_CONFIG.Blockchain.Ethereum {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepEthBlockchain(ctx)
		} else {
			testnet.SweepEthSepoliaBlockchain(ctx)
		}
	}

	if global.NODE_CONFIG.Blockchain.Bsc {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepBscBlockchain(ctx)
		} else {
			testnet.SweepBscTestnetBlockchain(ctx)
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
			mainnet.SweepOpBlockchain(ctx)
		} else {
			testnet.SweepOpSepoliaBlockchain(ctx)
		}
	}

	if global.NODE_CONFIG.Blockchain.ArbitrumOne {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepArbitrumOneBlockchain(ctx)
		} else {
			testnet.SweepArbitrumSepoliaBlockchain(ctx)
		}
	}

	if global.NODE_CONFIG.Blockchain.ArbitrumNova {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepArbitrumNovaBlockchain(ctx)
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
		} else {
			// testnet.SweepBchTestnetBlockchain(ctx)
		}
	}

	if global.NODE_CONFIG.Blockchain.Pol {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepPolBlockchain(ctx)
		} else {
			testnet.SweepPolTestnetBlockchain(ctx)
		}
	}

	if global.NODE_CONFIG.Blockchain.Avax {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepAvaxBlockchain(ctx)
		} else {
			testnet.SweepAvaxTestnetBlockchain(ctx)
		}
	}

	if global.NODE_CONFIG.Blockchain.Base {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepBaseBlockchain(ctx)
		} else {
			testnet.SweepBaseSepoliaBlockchain(ctx)
		}
	}
}
