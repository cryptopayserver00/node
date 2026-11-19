package setup

import "node/global/constant"

// ChainState 单链扫块运行时状态与 Redis key。
type ChainState struct {
	PublicKeyKey            string
	LatestKey               string
	CacheKey                string
	SweepKey                string
	PendingBlockKey         string
	PendingTransactionKey   string

	PublicKeys *[]string
	Latest     *int64
	Cache      *int64
	Sweep      *int64
}

var chainRegistry map[uint]*ChainState

func init() {
	chainRegistry = map[uint]*ChainState{
		constant.ETH_MAINNET: {
			PublicKeyKey: constant.ETH_PUBLIC_KEY, LatestKey: constant.ETH_LATEST_BLOCK,
			CacheKey: constant.ETH_CACHE_BLOCK, SweepKey: constant.ETH_SWEEP_BLOCK,
			PendingBlockKey: constant.ETH_PENDING_BLOCK, PendingTransactionKey: constant.ETH_PENDING_TRANSACTION,
			PublicKeys: &EthPublicKey, Latest: &EthLatestBlockHeight, Cache: &EthCacheBlockHeight, Sweep: &EthSweepBlockHeight,
		},
		constant.ETH_SEPOLIA: {
			PublicKeyKey: constant.ETH_SEPOLIA_PUBLIC_KEY, LatestKey: constant.ETH_SEPOLIA_LATEST_BLOCK,
			CacheKey: constant.ETH_SEPOLIA_CACHE_BLOCK, SweepKey: constant.ETH_SEPOLIA_SWEEP_BLOCK,
			PendingBlockKey: constant.ETH_SEPOLIA_PENDING_BLOCK, PendingTransactionKey: constant.ETH_SEPOLIA_PENDING_TRANSACTION,
			PublicKeys: &EthSepoliaPublicKey, Latest: &EthSepoliaLatestBlockHeight, Cache: &EthSepoliaCacheBlockHeight, Sweep: &EthSepoliaSweepBlockHeight,
		},
		constant.BTC_MAINNET: {
			PublicKeyKey: constant.BTC_PUBLIC_KEY, LatestKey: constant.BTC_LATEST_BLOCK,
			CacheKey: constant.BTC_CACHE_BLOCK, SweepKey: constant.BTC_SWEEP_BLOCK,
			PendingBlockKey: constant.BTC_PENDING_BLOCK, PendingTransactionKey: constant.BTC_PENDING_TRANSACTION,
			PublicKeys: &BtcPublicKey, Latest: &BtcLatestBlockHeight, Cache: &BtcCacheBlockHeight, Sweep: &BtcSweepBlockHeight,
		},
		constant.BTC_TESTNET: {
			PublicKeyKey: constant.BTC_TESTNET_PUBLIC_KEY, LatestKey: constant.BTC_TESTNET_LATEST_BLOCK,
			CacheKey: constant.BTC_TESTNET_CACHE_BLOCK, SweepKey: constant.BTC_TESTNET_SWEEP_BLOCK,
			PendingBlockKey: constant.BTC_TESTNET_PENDING_BLOCK, PendingTransactionKey: constant.BTC_TESTNET_PENDING_TRANSACTION,
			PublicKeys: &BtcTestnetPublicKey, Latest: &BtcTestnetLatestBlockHeight, Cache: &BtcTestnetCacheBlockHeight, Sweep: &BtcTestnetSweepBlockHeight,
		},
		constant.BSC_MAINNET: {
			PublicKeyKey: constant.BSC_PUBLIC_KEY, LatestKey: constant.BSC_LATEST_BLOCK,
			CacheKey: constant.BSC_CACHE_BLOCK, SweepKey: constant.BSC_SWEEP_BLOCK,
			PendingBlockKey: constant.BSC_PENDING_BLOCK, PendingTransactionKey: constant.BSC_PENDING_TRANSACTION,
			PublicKeys: &BscPublicKey, Latest: &BscLatestBlockHeight, Cache: &BscCacheBlockHeight, Sweep: &BscSweepBlockHeight,
		},
		constant.BSC_TESTNET: {
			PublicKeyKey: constant.BSC_TESTNET_PUBLIC_KEY, LatestKey: constant.BSC_TESTNET_LATEST_BLOCK,
			CacheKey: constant.BSC_TESTNET_CACHE_BLOCK, SweepKey: constant.BSC_TESTNET_SWEEP_BLOCK,
			PendingBlockKey: constant.BSC_TESTNET_PENDING_BLOCK, PendingTransactionKey: constant.BSC_TESTNET_PENDING_TRANSACTION,
			PublicKeys: &BscTestnetPublicKey, Latest: &BscTestnetLatestBlockHeight, Cache: &BscTestnetCacheBlockHeight, Sweep: &BscTestnetSweepBlockHeight,
		},
		constant.ARBITRUM_ONE: {
			PublicKeyKey: constant.ARBITRUM_ONE_PUBLIC_KEY, LatestKey: constant.ARBITRUM_ONE_LATEST_BLOCK,
			CacheKey: constant.ARBITRUM_ONE_CACHE_BLOCK, SweepKey: constant.ARBITRUM_ONE_SWEEP_BLOCK,
			PendingBlockKey: constant.ARBITRUM_ONE_PENDING_BLOCK, PendingTransactionKey: constant.ARBITRUM_ONE_PENDING_TRANSACTION,
			PublicKeys: &ArbitrumOnePublicKey, Latest: &ArbitrumOneLatestBlockHeight, Cache: &ArbitrumOneCacheBlockHeight, Sweep: &ArbitrumOneSweepBlockHeight,
		},
		constant.ARBITRUM_NOVA: {
			PublicKeyKey: constant.ARBITRUM_NOVA_PUBLIC_KEY, LatestKey: constant.ARBITRUM_NOVA_LATEST_BLOCK,
			CacheKey: constant.ARBITRUM_NOVA_CACHE_BLOCK, SweepKey: constant.ARBITRUM_NOVA_SWEEP_BLOCK,
			PendingBlockKey: constant.ARBITRUM_NOVA_PENDING_BLOCK, PendingTransactionKey: constant.ARBITRUM_NOVA_PENDING_TRANSACTION,
			PublicKeys: &ArbitrumNovaPublicKey, Latest: &ArbitrumNovaLatestBlockHeight, Cache: &ArbitrumNovaCacheBlockHeight, Sweep: &ArbitrumNovaSweepBlockHeight,
		},
		constant.ARBITRUM_SEPOLIA: {
			PublicKeyKey: constant.ARBITRUM_SEPOLIA_PUBLIC_KEY, LatestKey: constant.ARBITRUM_SEPOLIA_LATEST_BLOCK,
			CacheKey: constant.ARBITRUM_SEPOLIA_CACHE_BLOCK, SweepKey: constant.ARBITRUM_SEPOLIA_SWEEP_BLOCK,
			PendingBlockKey: constant.ARBITRUM_SEPOLIA_PENDING_BLOCK, PendingTransactionKey: constant.ARBITRUM_SEPOLIA_PENDING_TRANSACTION,
			PublicKeys: &ArbitrumSepoliaPublicKey, Latest: &ArbitrumSepoliaLatestBlockHeight, Cache: &ArbitrumSepoliaCacheBlockHeight, Sweep: &ArbitrumSepoliaSweepBlockHeight,
		},
		constant.TRON_MAINNET: {
			PublicKeyKey: constant.TRON_PUBLIC_KEY, LatestKey: constant.TRON_LATEST_BLOCK,
			CacheKey: constant.TRON_CACHE_BLOCK, SweepKey: constant.TRON_SWEEP_BLOCK,
			PendingBlockKey: constant.TRON_PENDING_BLOCK, PendingTransactionKey: constant.TRON_PENDING_TRANSACTION,
			PublicKeys: &TronPublicKey, Latest: &TronLatestBlockHeight, Cache: &TronCacheBlockHeight, Sweep: &TronSweepBlockHeight,
		},
		constant.TRON_NILE: {
			PublicKeyKey: constant.TRON_NILE_PUBLIC_KEY, LatestKey: constant.TRON_NILE_LATEST_BLOCK,
			CacheKey: constant.TRON_NILE_CACHE_BLOCK, SweepKey: constant.TRON_NILE_SWEEP_BLOCK,
			PendingBlockKey: constant.TRON_NILE_PENDING_BLOCK, PendingTransactionKey: constant.TRON_NILE_PENDING_TRANSACTION,
			PublicKeys: &TronNilePublicKey, Latest: &TronNileLatestBlockHeight, Cache: &TronNileCacheBlockHeight, Sweep: &TronNileSweepBlockHeight,
		},
		constant.LTC_MAINNET: {
			PublicKeyKey: constant.LTC_PUBLIC_KEY, LatestKey: constant.LTC_LATEST_BLOCK,
			CacheKey: constant.LTC_CACHE_BLOCK, SweepKey: constant.LTC_SWEEP_BLOCK,
			PendingBlockKey: constant.LTC_PENDING_BLOCK, PendingTransactionKey: constant.LTC_PENDING_TRANSACTION,
			PublicKeys: &LtcPublicKey, Latest: &LtcLatestBlockHeight, Cache: &LtcCacheBlockHeight, Sweep: &LtcSweepBlockHeight,
		},
		constant.LTC_TESTNET: {
			PublicKeyKey: constant.LTC_TESTNET_PUBLIC_KEY, LatestKey: constant.LTC_TESTNET_LATEST_BLOCK,
			CacheKey: constant.LTC_TESTNET_CACHE_BLOCK, SweepKey: constant.LTC_TESTNET_SWEEP_BLOCK,
			PendingBlockKey: constant.LTC_TESTNET_PENDING_BLOCK, PendingTransactionKey: constant.LTC_TESTNET_PENDING_TRANSACTION,
			PublicKeys: &LtcTestnetPublicKey, Latest: &LtcTestnetLatestBlockHeight, Cache: &LtcTestnetCacheBlockHeight, Sweep: &LtcTestnetSweepBlockHeight,
		},
		constant.OP_MAINNET: {
			PublicKeyKey: constant.OP_PUBLIC_KEY, LatestKey: constant.OP_LATEST_BLOCK,
			CacheKey: constant.OP_CACHE_BLOCK, SweepKey: constant.OP_SWEEP_BLOCK,
			PendingBlockKey: constant.OP_PENDING_BLOCK, PendingTransactionKey: constant.OP_PENDING_TRANSACTION,
			PublicKeys: &OpPublicKey, Latest: &OpLatestBlockHeight, Cache: &OpCacheBlockHeight, Sweep: &OpSweepBlockHeight,
		},
		constant.OP_SEPOLIA: {
			PublicKeyKey: constant.OP_SEPOLIA_PUBLIC_KEY, LatestKey: constant.OP_SEPOLIA_LATEST_BLOCK,
			CacheKey: constant.OP_SEPOLIA_CACHE_BLOCK, SweepKey: constant.OP_SEPOLIA_SWEEP_BLOCK,
			PendingBlockKey: constant.OP_SEPOLIA_PENDING_BLOCK, PendingTransactionKey: constant.OP_SEPOLIA_PENDING_TRANSACTION,
			PublicKeys: &OpSepoliaPublicKey, Latest: &OpSepoliaLatestBlockHeight, Cache: &OpSepoliaCacheBlockHeight, Sweep: &OpSepoliaSweepBlockHeight,
		},
		constant.SOL_MAINNET: {
			PublicKeyKey: constant.SOL_PUBLIC_KEY, LatestKey: constant.SOL_LATEST_BLOCK,
			CacheKey: constant.SOL_CACHE_BLOCK, SweepKey: constant.SOL_SWEEP_BLOCK,
			PendingBlockKey: constant.SOL_PENDING_BLOCK, PendingTransactionKey: constant.SOL_PENDING_TRANSACTION,
			PublicKeys: &SolPublicKey, Latest: &SolLatestBlockHeight, Cache: &SolCacheBlockHeight, Sweep: &SolSweepBlockHeight,
		},
		constant.SOL_DEVNET: {
			PublicKeyKey: constant.SOL_DEVNET_PUBLIC_KEY, LatestKey: constant.SOL_DEVNET_LATEST_BLOCK,
			CacheKey: constant.SOL_DEVNET_CACHE_BLOCK, SweepKey: constant.SOL_DEVNET_SWEEP_BLOCK,
			PendingBlockKey: constant.SOL_DEVNET_PENDING_BLOCK, PendingTransactionKey: constant.SOL_DEVNET_PENDING_TRANSACTION,
			PublicKeys: &SolDevnetPublicKey, Latest: &SolDevnetLatestBlockHeight, Cache: &SolDevnetCacheBlockHeight, Sweep: &SolDevnetSweepBlockHeight,
		},
		constant.TON_MAINNET: {
			PublicKeyKey: constant.TON_PUBLIC_KEY, LatestKey: constant.TON_LATEST_BLOCK,
			CacheKey: constant.TON_CACHE_BLOCK, SweepKey: constant.TON_SWEEP_BLOCK,
			PendingBlockKey: constant.TON_PENDING_BLOCK, PendingTransactionKey: constant.TON_PENDING_TRANSACTION,
			PublicKeys: &TonPublicKey, Latest: &TonLatestBlockHeight, Cache: &TonCacheBlockHeight, Sweep: &TonSweepBlockHeight,
		},
		constant.TON_TESTNET: {
			PublicKeyKey: constant.TON_TESTNET_PUBLIC_KEY, LatestKey: constant.TON_TESTNET_LATEST_BLOCK,
			CacheKey: constant.TON_TESTNET_CACHE_BLOCK, SweepKey: constant.TON_TESTNET_SWEEP_BLOCK,
			PendingBlockKey: constant.TON_TESTNET_PENDING_BLOCK, PendingTransactionKey: constant.TON_TESTNET_PENDING_TRANSACTION,
			PublicKeys: &TonTestnetPublicKey, Latest: &TonTestnetLatestBlockHeight, Cache: &TonTestnetCacheBlockHeight, Sweep: &TonTestnetSweepBlockHeight,
		},
		constant.XRP_MAINNET: {
			PublicKeyKey: constant.XRP_PUBLIC_KEY, LatestKey: constant.XRP_LATEST_BLOCK,
			CacheKey: constant.XRP_CACHE_BLOCK, SweepKey: constant.XRP_SWEEP_BLOCK,
			PendingBlockKey: constant.XRP_PENDING_BLOCK, PendingTransactionKey: constant.XRP_PENDING_TRANSACTION,
			PublicKeys: &XrpPublicKey, Latest: &XrpLatestBlockHeight, Cache: &XrpCacheBlockHeight, Sweep: &XrpSweepBlockHeight,
		},
		constant.XRP_TESTNET: {
			PublicKeyKey: constant.XRP_TESTNET_PUBLIC_KEY, LatestKey: constant.XRP_TESTNET_LATEST_BLOCK,
			CacheKey: constant.XRP_TESTNET_CACHE_BLOCK, SweepKey: constant.XRP_TESTNET_SWEEP_BLOCK,
			PendingBlockKey: constant.XRP_TESTNET_PENDING_BLOCK, PendingTransactionKey: constant.XRP_TESTNET_PENDING_TRANSACTION,
			PublicKeys: &XrpTestnetPublicKey, Latest: &XrpTestnetLatestBlockHeight, Cache: &XrpTestnetCacheBlockHeight, Sweep: &XrpTestnetSweepBlockHeight,
		},
		constant.BCH_MAINNET: {
			PublicKeyKey: constant.BCH_PUBLIC_KEY, LatestKey: constant.BCH_LATEST_BLOCK,
			CacheKey: constant.BCH_CACHE_BLOCK, SweepKey: constant.BCH_SWEEP_BLOCK,
			PendingBlockKey: constant.BCH_PENDING_BLOCK, PendingTransactionKey: constant.BCH_PENDING_TRANSACTION,
			PublicKeys: &BchPublicKey, Latest: &BchLatestBlockHeight, Cache: &BchCacheBlockHeight, Sweep: &BchSweepBlockHeight,
		},
		constant.BCH_TESTNET: {
			PublicKeyKey: constant.BCH_TESTNET_PUBLIC_KEY, LatestKey: constant.BCH_TESTNET_LATEST_BLOCK,
			CacheKey: constant.BCH_TESTNET_CACHE_BLOCK, SweepKey: constant.BCH_TESTNET_SWEEP_BLOCK,
			PendingBlockKey: constant.BCH_TESTNET_PENDING_BLOCK, PendingTransactionKey: constant.BCH_TESTNET_PENDING_TRANSACTION,
			PublicKeys: &BchTestnetPublicKey, Latest: &BchTestnetLatestBlockHeight, Cache: &BchTestnetCacheBlockHeight, Sweep: &BchTestnetSweepBlockHeight,
		},
		constant.POL_MAINNET: {
			PublicKeyKey: constant.POL_PUBLIC_KEY, LatestKey: constant.POL_LATEST_BLOCK,
			CacheKey: constant.POL_CACHE_BLOCK, SweepKey: constant.POL_SWEEP_BLOCK,
			PendingBlockKey: constant.POL_PENDING_BLOCK, PendingTransactionKey: constant.POL_PENDING_TRANSACTION,
			PublicKeys: &PolPublicKey, Latest: &PolLatestBlockHeight, Cache: &PolCacheBlockHeight, Sweep: &PolSweepBlockHeight,
		},
		constant.POL_TESTNET: {
			PublicKeyKey: constant.POL_TESTNET_PUBLIC_KEY, LatestKey: constant.POL_TESTNET_LATEST_BLOCK,
			CacheKey: constant.POL_TESTNET_CACHE_BLOCK, SweepKey: constant.POL_TESTNET_SWEEP_BLOCK,
			PendingBlockKey: constant.POL_TESTNET_PENDING_BLOCK, PendingTransactionKey: constant.POL_TESTNET_PENDING_TRANSACTION,
			PublicKeys: &PolTestnetPublicKey, Latest: &PolTestnetLatestBlockHeight, Cache: &PolTestnetCacheBlockHeight, Sweep: &PolTestnetSweepBlockHeight,
		},
		constant.AVAX_MAINNET: {
			PublicKeyKey: constant.AVAX_PUBLIC_KEY, LatestKey: constant.AVAX_LATEST_BLOCK,
			CacheKey: constant.AVAX_CACHE_BLOCK, SweepKey: constant.AVAX_SWEEP_BLOCK,
			PendingBlockKey: constant.AVAX_PENDING_BLOCK, PendingTransactionKey: constant.AVAX_PENDING_TRANSACTION,
			PublicKeys: &AvaxPublicKey, Latest: &AvaxLatestBlockHeight, Cache: &AvaxCacheBlockHeight, Sweep: &AvaxSweepBlockHeight,
		},
		constant.AVAX_TESTNET: {
			PublicKeyKey: constant.AVAX_TESTNET_PUBLIC_KEY, LatestKey: constant.AVAX_TESTNET_LATEST_BLOCK,
			CacheKey: constant.AVAX_TESTNET_CACHE_BLOCK, SweepKey: constant.AVAX_TESTNET_SWEEP_BLOCK,
			PendingBlockKey: constant.AVAX_TESTNET_PENDING_BLOCK, PendingTransactionKey: constant.AVAX_TESTNET_PENDING_TRANSACTION,
			PublicKeys: &AvaxTestnetPublicKey, Latest: &AvaxTestnetLatestBlockHeight, Cache: &AvaxTestnetCacheBlockHeight, Sweep: &AvaxTestnetSweepBlockHeight,
		},
		constant.BASE_MAINNET: {
			PublicKeyKey: constant.BASE_PUBLIC_KEY, LatestKey: constant.BASE_LATEST_BLOCK,
			CacheKey: constant.BASE_CACHE_BLOCK, SweepKey: constant.BASE_SWEEP_BLOCK,
			PendingBlockKey: constant.BASE_PENDING_BLOCK, PendingTransactionKey: constant.BASE_PENDING_TRANSACTION,
			PublicKeys: &BasePublicKey, Latest: &BaseLatestBlockHeight, Cache: &BaseCacheBlockHeight, Sweep: &BaseSweepBlockHeight,
		},
		constant.BASE_SEPOLIA: {
			PublicKeyKey: constant.BASE_SEPOLIA_PUBLIC_KEY, LatestKey: constant.BASE_SEPOLIA_LATEST_BLOCK,
			CacheKey: constant.BASE_SEPOLIA_CACHE_BLOCK, SweepKey: constant.BASE_SEPOLIA_SWEEP_BLOCK,
			PendingBlockKey: constant.BASE_SEPOLIA_PENDING_BLOCK, PendingTransactionKey: constant.BASE_SEPOLIA_PENDING_TRANSACTION,
			PublicKeys: &BaseSepoliaPublicKey, Latest: &BaseSepoliaLatestBlockHeight, Cache: &BaseSepoliaCacheBlockHeight, Sweep: &BaseSepoliaSweepBlockHeight,
		},
	}
}

func GetChainState(chainId uint) (*ChainState, bool) {
	st, ok := chainRegistry[chainId]
	return st, ok
}
