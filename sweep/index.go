package sweep

import (
	"context"
	"node/global"
	"node/sweep/mainnet"
	"node/sweep/setup"
	"node/sweep/testnet"
)

func RunBlockSweep() {
	setup.SetupPublicKey(context.Background())

	if global.NODE_CONFIG.Blockchain.Ethereum {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepEthBlockchain()
		} else {
			testnet.SweepEthSepoliaBlockchain()
		}
	}

	if global.NODE_CONFIG.Blockchain.Bsc {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepBscBlockchain()
		} else {
			testnet.SweepBscTestnetBlockchain()
		}

	}

	if global.NODE_CONFIG.Blockchain.Bitcoin {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepBtcBlockchain()
		} else {
			testnet.SweepBtcTestnetBlockchain()
		}
	}

	if global.NODE_CONFIG.Blockchain.Tron {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepTronBlockchain()
		} else {
			testnet.SweepTronNileBlockchain()
		}
	}

	if global.NODE_CONFIG.Blockchain.Litecoin {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepLtcBlockchain()
		} else {
			testnet.SweepLtcTestnetBlockchain()
		}
	}

	if global.NODE_CONFIG.Blockchain.Op {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepOpBlockchain()
		} else {
			testnet.SweepOpSepoliaBlockchain()
		}
	}

	if global.NODE_CONFIG.Blockchain.ArbitrumOne {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepArbitrumOneBlockchain()
		} else {
			// testnet.SweepArbitrumGoerliBlockchain()
			testnet.SweepArbitrumSepoliaBlockchain()
		}
	}

	if global.NODE_CONFIG.Blockchain.ArbitrumNova {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepArbitrumNovaBlockchain()
		} else {
		}
	}

	if global.NODE_CONFIG.Blockchain.Solana {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepSolBlockchain()
		} else {
			testnet.SweepSolDevnetBlockchain()
		}
	}

	if global.NODE_CONFIG.Blockchain.Ton {
		if global.NODE_CONFIG.Blockchain.SweepMainnet {
			mainnet.SweepTonBlockchain()
		} else {
			testnet.SweepTonTestnetBlockchain()
		}
	}
}
