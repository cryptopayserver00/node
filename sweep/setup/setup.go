package setup

import (
	"context"
	"errors"
	"fmt"
	"node/global"
	"node/global/constant"
	"node/model"
	"node/sweep/utils"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var (
	SweepThreshold = 5

	SweepPublicKeyArray = []string{
		constant.ETH_PUBLIC_KEY,
		constant.ETH_SEPOLIA_PUBLIC_KEY,
		constant.BTC_PUBLIC_KEY,
		constant.BTC_TESTNET_PUBLIC_KEY,
		constant.TRON_PUBLIC_KEY,
		constant.TRON_NILE_PUBLIC_KEY,
		constant.BSC_PUBLIC_KEY,
		constant.BSC_TESTNET_PUBLIC_KEY,
		constant.ARBITRUM_ONE_PUBLIC_KEY,
		constant.ARBITRUM_NOVA_PUBLIC_KEY,
		constant.ARBITRUM_SEPOLIA_PUBLIC_KEY,
		constant.LTC_PUBLIC_KEY,
		constant.LTC_TESTNET_PUBLIC_KEY,
		constant.OP_PUBLIC_KEY,
		constant.OP_SEPOLIA_PUBLIC_KEY,
		constant.SOL_PUBLIC_KEY,
		constant.SOL_DEVNET_PUBLIC_KEY,
		constant.TON_PUBLIC_KEY,
		constant.TON_TESTNET_PUBLIC_KEY,
		constant.XRP_PUBLIC_KEY,
		constant.XRP_TESTNET_PUBLIC_KEY,
		constant.BCH_PUBLIC_KEY,
		constant.BCH_TESTNET_PUBLIC_KEY,
		constant.POL_PUBLIC_KEY,
		constant.POL_TESTNET_PUBLIC_KEY,
		constant.AVAX_PUBLIC_KEY,
		constant.AVAX_TESTNET_PUBLIC_KEY,
		constant.BASE_PUBLIC_KEY,
		constant.BASE_SEPOLIA_PUBLIC_KEY,
	}

	EthPublicKey             []string
	EthSepoliaPublicKey      []string
	BtcPublicKey             []string
	BtcTestnetPublicKey      []string
	BscPublicKey             []string
	BscTestnetPublicKey      []string
	ArbitrumOnePublicKey     []string
	ArbitrumNovaPublicKey    []string
	ArbitrumSepoliaPublicKey []string
	TronPublicKey            []string
	TronNilePublicKey        []string
	LtcPublicKey             []string
	LtcTestnetPublicKey      []string
	OpPublicKey              []string
	OpSepoliaPublicKey       []string
	SolPublicKey             []string
	SolDevnetPublicKey       []string
	TonPublicKey             []string
	TonTestnetPublicKey      []string
	XrpPublicKey             []string
	XrpTestnetPublicKey      []string
	BchPublicKey             []string
	BchTestnetPublicKey      []string
	PolPublicKey             []string
	PolTestnetPublicKey      []string
	AvaxPublicKey            []string
	AvaxTestnetPublicKey     []string
	BasePublicKey            []string
	BaseSepoliaPublicKey     []string

	EthLatestBlockHeight int64
	EthCacheBlockHeight  int64
	EthSweepBlockHeight  int64

	EthSepoliaLatestBlockHeight int64
	EthSepoliaCacheBlockHeight  int64
	EthSepoliaSweepBlockHeight  int64

	BtcLatestBlockHeight int64
	BtcCacheBlockHeight  int64
	BtcSweepBlockHeight  int64

	BtcTestnetLatestBlockHeight int64
	BtcTestnetCacheBlockHeight  int64
	BtcTestnetSweepBlockHeight  int64

	BscLatestBlockHeight int64
	BscCacheBlockHeight  int64
	BscSweepBlockHeight  int64

	BscTestnetLatestBlockHeight int64
	BscTestnetCacheBlockHeight  int64
	BscTestnetSweepBlockHeight  int64

	ArbitrumOneLatestBlockHeight int64
	ArbitrumOneCacheBlockHeight  int64
	ArbitrumOneSweepBlockHeight  int64

	ArbitrumNovaLatestBlockHeight int64
	ArbitrumNovaCacheBlockHeight  int64
	ArbitrumNovaSweepBlockHeight  int64

	ArbitrumSepoliaLatestBlockHeight int64
	ArbitrumSepoliaCacheBlockHeight  int64
	ArbitrumSepoliaSweepBlockHeight  int64

	TronLatestBlockHeight int64
	TronCacheBlockHeight  int64
	TronSweepBlockHeight  int64

	TronNileLatestBlockHeight int64
	TronNileCacheBlockHeight  int64
	TronNileSweepBlockHeight  int64

	LtcLatestBlockHeight int64
	LtcCacheBlockHeight  int64
	LtcSweepBlockHeight  int64

	LtcTestnetLatestBlockHeight int64
	LtcTestnetCacheBlockHeight  int64
	LtcTestnetSweepBlockHeight  int64

	OpLatestBlockHeight int64
	OpCacheBlockHeight  int64
	OpSweepBlockHeight  int64

	OpSepoliaLatestBlockHeight int64
	OpSepoliaCacheBlockHeight  int64
	OpSepoliaSweepBlockHeight  int64

	SolLatestBlockHeight int64
	SolCacheBlockHeight  int64
	SolSweepBlockHeight  int64

	SolDevnetLatestBlockHeight int64
	SolDevnetCacheBlockHeight  int64
	SolDevnetSweepBlockHeight  int64

	TonLatestBlockHeight int64
	TonCacheBlockHeight  int64
	TonSweepBlockHeight  int64

	TonTestnetLatestBlockHeight int64
	TonTestnetCacheBlockHeight  int64
	TonTestnetSweepBlockHeight  int64

	XrpLatestBlockHeight int64
	XrpCacheBlockHeight  int64
	XrpSweepBlockHeight  int64

	XrpTestnetLatestBlockHeight int64
	XrpTestnetCacheBlockHeight  int64
	XrpTestnetSweepBlockHeight  int64

	BchLatestBlockHeight int64
	BchCacheBlockHeight  int64
	BchSweepBlockHeight  int64

	BchTestnetLatestBlockHeight int64
	BchTestnetCacheBlockHeight  int64
	BchTestnetSweepBlockHeight  int64

	PolLatestBlockHeight int64
	PolCacheBlockHeight  int64
	PolSweepBlockHeight  int64

	PolTestnetLatestBlockHeight int64
	PolTestnetCacheBlockHeight  int64
	PolTestnetSweepBlockHeight  int64

	AvaxLatestBlockHeight int64
	AvaxCacheBlockHeight  int64
	AvaxSweepBlockHeight  int64

	AvaxTestnetLatestBlockHeight int64
	AvaxTestnetCacheBlockHeight  int64
	AvaxTestnetSweepBlockHeight  int64

	BaseLatestBlockHeight int64
	BaseCacheBlockHeight  int64
	BaseSweepBlockHeight  int64

	BaseSepoliaLatestBlockHeight int64
	BaseSepoliaCacheBlockHeight  int64
	BaseSepoliaSweepBlockHeight  int64
)

func SetupPublicKey(ctx context.Context) {
	var err error
	for _, v := range SweepPublicKeyArray {
		_, err = global.NODE_REDIS.Del(ctx, v).Result()
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			return
		}
	}

	var wallets []model.Wallet
	err = global.NODE_DB.Select("chain_id", "address").Find(&wallets).Error
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	if len(wallets) > 0 {
		for _, w := range wallets {
			switch w.ChainId {
			case constant.ETH_MAINNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.ETH_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				EthPublicKey = append(EthPublicKey, w.Address)
			case constant.ETH_SEPOLIA:
				_, err = global.NODE_REDIS.RPush(ctx, constant.ETH_SEPOLIA_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				EthSepoliaPublicKey = append(EthSepoliaPublicKey, w.Address)
			case constant.BTC_MAINNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.BTC_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				BtcPublicKey = append(BtcPublicKey, w.Address)
			case constant.BTC_TESTNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.BTC_TESTNET_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				BtcTestnetPublicKey = append(BtcTestnetPublicKey, w.Address)
			case constant.BSC_MAINNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.BSC_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				BscPublicKey = append(BscPublicKey, w.Address)
			case constant.BSC_TESTNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.BSC_TESTNET_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				BscTestnetPublicKey = append(BscTestnetPublicKey, w.Address)
			case constant.ARBITRUM_ONE:
				_, err = global.NODE_REDIS.RPush(ctx, constant.ARBITRUM_ONE_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				ArbitrumOnePublicKey = append(ArbitrumOnePublicKey, w.Address)
			case constant.ARBITRUM_NOVA:
				_, err = global.NODE_REDIS.RPush(ctx, constant.ARBITRUM_NOVA_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				ArbitrumNovaPublicKey = append(ArbitrumNovaPublicKey, w.Address)
			case constant.ARBITRUM_SEPOLIA:
				_, err = global.NODE_REDIS.RPush(ctx, constant.ARBITRUM_SEPOLIA_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				ArbitrumSepoliaPublicKey = append(ArbitrumSepoliaPublicKey, w.Address)
			case constant.TRON_MAINNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.TRON_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				TronPublicKey = append(TronPublicKey, w.Address)
			case constant.TRON_NILE:
				_, err = global.NODE_REDIS.RPush(ctx, constant.TRON_NILE_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				TronNilePublicKey = append(TronNilePublicKey, w.Address)
			case constant.LTC_MAINNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.LTC_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				LtcPublicKey = append(LtcPublicKey, w.Address)
			case constant.LTC_TESTNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.LTC_TESTNET_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				LtcTestnetPublicKey = append(LtcTestnetPublicKey, w.Address)
			case constant.OP_MAINNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.OP_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				OpPublicKey = append(OpPublicKey, w.Address)
			case constant.OP_SEPOLIA:
				_, err = global.NODE_REDIS.RPush(ctx, constant.OP_SEPOLIA_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				OpSepoliaPublicKey = append(OpSepoliaPublicKey, w.Address)
			case constant.SOL_MAINNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.SOL_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				SolPublicKey = append(SolPublicKey, w.Address)
			case constant.SOL_DEVNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.SOL_DEVNET_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				SolDevnetPublicKey = append(SolDevnetPublicKey, w.Address)
			case constant.TON_MAINNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.TON_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				TonPublicKey = append(TonPublicKey, w.Address)
			case constant.TON_TESTNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.TON_TESTNET_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				TonTestnetPublicKey = append(TonTestnetPublicKey, w.Address)
			case constant.XRP_MAINNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.XRP_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				XrpPublicKey = append(XrpPublicKey, w.Address)
			case constant.XRP_TESTNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.XRP_TESTNET_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				XrpTestnetPublicKey = append(XrpTestnetPublicKey, w.Address)
			case constant.BCH_MAINNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.BCH_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				BchPublicKey = append(BchPublicKey, w.Address)
			case constant.BCH_TESTNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.BCH_TESTNET_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				BchTestnetPublicKey = append(BchTestnetPublicKey, w.Address)
			case constant.POL_MAINNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.POL_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				PolPublicKey = append(PolPublicKey, w.Address)
			case constant.POL_TESTNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.POL_TESTNET_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				PolTestnetPublicKey = append(PolTestnetPublicKey, w.Address)
			case constant.AVAX_MAINNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.AVAX_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				AvaxPublicKey = append(AvaxPublicKey, w.Address)
			case constant.AVAX_TESTNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.AVAX_TESTNET_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				AvaxTestnetPublicKey = append(AvaxTestnetPublicKey, w.Address)
			case constant.BASE_MAINNET:
				_, err = global.NODE_REDIS.RPush(ctx, constant.BASE_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				BasePublicKey = append(BasePublicKey, w.Address)
			case constant.BASE_SEPOLIA:
				_, err = global.NODE_REDIS.RPush(ctx, constant.BASE_SEPOLIA_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				BaseSepoliaPublicKey = append(BaseSepoliaPublicKey, w.Address)
			}
		}
	}
}

func SetupLatestBlockHeight(ctx context.Context, chainId uint, blockNumber int64) {
	if !utils.IsChainJoinSweep(chainId) {
		return
	}

	var err error
	var latestBlockKey string
	var latestHeightVal *int64

	switch chainId {
	case constant.ETH_MAINNET:
		latestBlockKey = constant.ETH_LATEST_BLOCK
		latestHeightVal = &EthLatestBlockHeight
	case constant.ETH_SEPOLIA:
		latestBlockKey = constant.ETH_SEPOLIA_LATEST_BLOCK
		latestHeightVal = &EthSepoliaLatestBlockHeight
	case constant.BTC_MAINNET:
		latestBlockKey = constant.BTC_LATEST_BLOCK
		latestHeightVal = &BtcLatestBlockHeight
	case constant.BTC_TESTNET:
		latestBlockKey = constant.BTC_TESTNET_LATEST_BLOCK
		latestHeightVal = &BtcTestnetLatestBlockHeight
	case constant.BSC_MAINNET:
		latestBlockKey = constant.BSC_LATEST_BLOCK
		latestHeightVal = &BscLatestBlockHeight
	case constant.BSC_TESTNET:
		latestBlockKey = constant.BSC_TESTNET_LATEST_BLOCK
		latestHeightVal = &BscTestnetLatestBlockHeight
	case constant.ARBITRUM_ONE:
		latestBlockKey = constant.ARBITRUM_ONE_LATEST_BLOCK
		latestHeightVal = &ArbitrumOneLatestBlockHeight
	case constant.ARBITRUM_NOVA:
		latestBlockKey = constant.ARBITRUM_NOVA_LATEST_BLOCK
		latestHeightVal = &ArbitrumNovaLatestBlockHeight
	case constant.ARBITRUM_SEPOLIA:
		latestBlockKey = constant.ARBITRUM_SEPOLIA_LATEST_BLOCK
		latestHeightVal = &ArbitrumSepoliaLatestBlockHeight
	case constant.TRON_MAINNET:
		latestBlockKey = constant.TRON_LATEST_BLOCK
		latestHeightVal = &TronLatestBlockHeight
	case constant.TRON_NILE:
		latestBlockKey = constant.TRON_NILE_LATEST_BLOCK
		latestHeightVal = &TronNileLatestBlockHeight
	case constant.LTC_MAINNET:
		latestBlockKey = constant.LTC_LATEST_BLOCK
		latestHeightVal = &LtcLatestBlockHeight
	case constant.LTC_TESTNET:
		latestBlockKey = constant.LTC_TESTNET_LATEST_BLOCK
		latestHeightVal = &LtcTestnetLatestBlockHeight
	case constant.OP_MAINNET:
		latestBlockKey = constant.OP_LATEST_BLOCK
		latestHeightVal = &OpLatestBlockHeight
	case constant.OP_SEPOLIA:
		latestBlockKey = constant.OP_SEPOLIA_LATEST_BLOCK
		latestHeightVal = &OpSepoliaLatestBlockHeight
	case constant.SOL_MAINNET:
		latestBlockKey = constant.SOL_LATEST_BLOCK
		latestHeightVal = &SolLatestBlockHeight
	case constant.SOL_DEVNET:
		latestBlockKey = constant.SOL_DEVNET_LATEST_BLOCK
		latestHeightVal = &SolDevnetLatestBlockHeight
	case constant.TON_MAINNET:
		latestBlockKey = constant.TON_LATEST_BLOCK
		latestHeightVal = &TonLatestBlockHeight
	case constant.TON_TESTNET:
		latestBlockKey = constant.TON_TESTNET_LATEST_BLOCK
		latestHeightVal = &TonTestnetLatestBlockHeight
	case constant.XRP_MAINNET:
		latestBlockKey = constant.XRP_LATEST_BLOCK
		latestHeightVal = &XrpLatestBlockHeight
	case constant.XRP_TESTNET:
		latestBlockKey = constant.XRP_TESTNET_LATEST_BLOCK
		latestHeightVal = &XrpTestnetLatestBlockHeight
	case constant.BCH_MAINNET:
		latestBlockKey = constant.BCH_LATEST_BLOCK
		latestHeightVal = &BchLatestBlockHeight
	case constant.BCH_TESTNET:
		latestBlockKey = constant.BCH_TESTNET_LATEST_BLOCK
		latestHeightVal = &BchTestnetLatestBlockHeight
	case constant.POL_MAINNET:
		latestBlockKey = constant.POL_LATEST_BLOCK
		latestHeightVal = &PolLatestBlockHeight
	case constant.POL_TESTNET:
		latestBlockKey = constant.POL_TESTNET_LATEST_BLOCK
		latestHeightVal = &PolTestnetLatestBlockHeight
	case constant.AVAX_MAINNET:
		latestBlockKey = constant.AVAX_LATEST_BLOCK
		latestHeightVal = &AvaxLatestBlockHeight
	case constant.AVAX_TESTNET:
		latestBlockKey = constant.AVAX_TESTNET_LATEST_BLOCK
		latestHeightVal = &AvaxTestnetLatestBlockHeight
	case constant.BASE_MAINNET:
		latestBlockKey = constant.BASE_LATEST_BLOCK
		latestHeightVal = &BaseLatestBlockHeight
	case constant.BASE_SEPOLIA:
		latestBlockKey = constant.BASE_SEPOLIA_LATEST_BLOCK
		latestHeightVal = &BaseSepoliaLatestBlockHeight

	default:
		return
	}

	_, err = global.NODE_REDIS.Set(ctx, latestBlockKey, blockNumber, 0).Result()
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}
	*latestHeightVal = blockNumber
}

func SetupCacheBlockHeight(ctx context.Context, chainId uint) {
	if !utils.IsChainJoinSweep(chainId) {
		return
	}

	var err error
	var cacheBlockKey string
	var cacheHeightVal *int64
	var latestBlockHeight int64

	switch chainId {
	case constant.ETH_MAINNET:
		cacheBlockKey = constant.ETH_CACHE_BLOCK
		cacheHeightVal = &EthCacheBlockHeight
		latestBlockHeight = EthLatestBlockHeight
	case constant.ETH_SEPOLIA:
		cacheBlockKey = constant.ETH_SEPOLIA_CACHE_BLOCK
		cacheHeightVal = &EthSepoliaCacheBlockHeight
		latestBlockHeight = EthSepoliaLatestBlockHeight
	case constant.BTC_MAINNET:
		cacheBlockKey = constant.BTC_CACHE_BLOCK
		cacheHeightVal = &BtcCacheBlockHeight
		latestBlockHeight = BtcLatestBlockHeight
	case constant.BTC_TESTNET:
		cacheBlockKey = constant.BTC_TESTNET_CACHE_BLOCK
		cacheHeightVal = &BtcTestnetCacheBlockHeight
		latestBlockHeight = BtcTestnetLatestBlockHeight
	case constant.BSC_MAINNET:
		cacheBlockKey = constant.BSC_CACHE_BLOCK
		cacheHeightVal = &BscCacheBlockHeight
		latestBlockHeight = BscLatestBlockHeight
	case constant.BSC_TESTNET:
		cacheBlockKey = constant.BSC_TESTNET_CACHE_BLOCK
		cacheHeightVal = &BscTestnetCacheBlockHeight
		latestBlockHeight = BscTestnetLatestBlockHeight
	case constant.ARBITRUM_ONE:
		cacheBlockKey = constant.ARBITRUM_ONE_CACHE_BLOCK
		cacheHeightVal = &ArbitrumOneCacheBlockHeight
		latestBlockHeight = ArbitrumOneLatestBlockHeight
	case constant.ARBITRUM_NOVA:
		cacheBlockKey = constant.ARBITRUM_NOVA_CACHE_BLOCK
		cacheHeightVal = &ArbitrumNovaCacheBlockHeight
		latestBlockHeight = ArbitrumNovaLatestBlockHeight
	case constant.ARBITRUM_SEPOLIA:
		cacheBlockKey = constant.ARBITRUM_SEPOLIA_CACHE_BLOCK
		cacheHeightVal = &ArbitrumSepoliaCacheBlockHeight
		latestBlockHeight = ArbitrumSepoliaLatestBlockHeight
	case constant.TRON_MAINNET:
		cacheBlockKey = constant.TRON_CACHE_BLOCK
		cacheHeightVal = &TronCacheBlockHeight
		latestBlockHeight = TronLatestBlockHeight
	case constant.TRON_NILE:
		cacheBlockKey = constant.TRON_NILE_CACHE_BLOCK
		cacheHeightVal = &TronNileCacheBlockHeight
		latestBlockHeight = TronNileLatestBlockHeight
	case constant.LTC_MAINNET:
		cacheBlockKey = constant.LTC_CACHE_BLOCK
		cacheHeightVal = &LtcCacheBlockHeight
		latestBlockHeight = LtcLatestBlockHeight
	case constant.LTC_TESTNET:
		cacheBlockKey = constant.LTC_TESTNET_CACHE_BLOCK
		cacheHeightVal = &LtcTestnetCacheBlockHeight
		latestBlockHeight = LtcTestnetLatestBlockHeight
	case constant.OP_MAINNET:
		cacheBlockKey = constant.OP_CACHE_BLOCK
		cacheHeightVal = &OpCacheBlockHeight
		latestBlockHeight = OpLatestBlockHeight
	case constant.OP_SEPOLIA:
		cacheBlockKey = constant.OP_SEPOLIA_CACHE_BLOCK
		cacheHeightVal = &OpSepoliaCacheBlockHeight
		latestBlockHeight = OpSepoliaLatestBlockHeight
	case constant.SOL_MAINNET:
		cacheBlockKey = constant.SOL_CACHE_BLOCK
		cacheHeightVal = &SolCacheBlockHeight
		latestBlockHeight = SolLatestBlockHeight
	case constant.SOL_DEVNET:
		cacheBlockKey = constant.SOL_DEVNET_CACHE_BLOCK
		cacheHeightVal = &SolDevnetCacheBlockHeight
		latestBlockHeight = SolDevnetLatestBlockHeight
	case constant.TON_MAINNET:
		cacheBlockKey = constant.TON_CACHE_BLOCK
		cacheHeightVal = &TonCacheBlockHeight
		latestBlockHeight = TonLatestBlockHeight
	case constant.TON_TESTNET:
		cacheBlockKey = constant.TON_TESTNET_CACHE_BLOCK
		cacheHeightVal = &TonTestnetCacheBlockHeight
		latestBlockHeight = TonTestnetLatestBlockHeight
	case constant.XRP_MAINNET:
		cacheBlockKey = constant.XRP_CACHE_BLOCK
		cacheHeightVal = &XrpCacheBlockHeight
		latestBlockHeight = XrpLatestBlockHeight
	case constant.XRP_TESTNET:
		cacheBlockKey = constant.XRP_TESTNET_CACHE_BLOCK
		cacheHeightVal = &XrpTestnetCacheBlockHeight
		latestBlockHeight = XrpTestnetLatestBlockHeight
	case constant.BCH_MAINNET:
		cacheBlockKey = constant.BCH_CACHE_BLOCK
		cacheHeightVal = &BchCacheBlockHeight
		latestBlockHeight = BchLatestBlockHeight
	case constant.BCH_TESTNET:
		cacheBlockKey = constant.BCH_TESTNET_CACHE_BLOCK
		cacheHeightVal = &BchTestnetCacheBlockHeight
		latestBlockHeight = BchTestnetLatestBlockHeight
	case constant.POL_MAINNET:
		cacheBlockKey = constant.POL_CACHE_BLOCK
		cacheHeightVal = &PolCacheBlockHeight
		latestBlockHeight = PolLatestBlockHeight
	case constant.POL_TESTNET:
		cacheBlockKey = constant.POL_TESTNET_CACHE_BLOCK
		cacheHeightVal = &PolTestnetCacheBlockHeight
		latestBlockHeight = PolTestnetLatestBlockHeight
	case constant.AVAX_MAINNET:
		cacheBlockKey = constant.AVAX_CACHE_BLOCK
		cacheHeightVal = &AvaxCacheBlockHeight
		latestBlockHeight = AvaxLatestBlockHeight
	case constant.AVAX_TESTNET:
		cacheBlockKey = constant.AVAX_TESTNET_CACHE_BLOCK
		cacheHeightVal = &AvaxTestnetCacheBlockHeight
		latestBlockHeight = AvaxTestnetLatestBlockHeight
	case constant.BASE_MAINNET:
		cacheBlockKey = constant.BASE_CACHE_BLOCK
		cacheHeightVal = &BaseCacheBlockHeight
		latestBlockHeight = BaseLatestBlockHeight
	case constant.BASE_SEPOLIA:
		cacheBlockKey = constant.BASE_SEPOLIA_CACHE_BLOCK
		cacheHeightVal = &BaseSepoliaCacheBlockHeight
		latestBlockHeight = BaseSepoliaLatestBlockHeight
	default:
		return
	}

	cacheBlockHeightString, err := global.NODE_REDIS.Get(ctx, cacheBlockKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			*cacheHeightVal = latestBlockHeight
		} else {
			global.NODE_LOG.Error(err.Error())
			return
		}
	} else {
		*cacheHeightVal, err = strconv.ParseInt(cacheBlockHeightString, 10, 64)
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			return
		}
	}
}

func SetupSweepBlockHeight(ctx context.Context, chainId uint) {
	if !utils.IsChainJoinSweep(chainId) {
		return
	}

	var err error
	var sweepBlockKey string
	var sweepHeightVal *int64
	var cacheBlockHeight int64

	switch chainId {
	case constant.ETH_MAINNET:
		sweepBlockKey = constant.ETH_SWEEP_BLOCK
		sweepHeightVal = &EthSweepBlockHeight
		cacheBlockHeight = EthCacheBlockHeight
	case constant.ETH_SEPOLIA:
		sweepBlockKey = constant.ETH_SEPOLIA_SWEEP_BLOCK
		sweepHeightVal = &EthSepoliaSweepBlockHeight
		cacheBlockHeight = EthSepoliaCacheBlockHeight
	case constant.BTC_MAINNET:
		sweepBlockKey = constant.BTC_SWEEP_BLOCK
		sweepHeightVal = &BtcSweepBlockHeight
		cacheBlockHeight = BtcCacheBlockHeight
	case constant.BTC_TESTNET:
		sweepBlockKey = constant.BTC_TESTNET_SWEEP_BLOCK
		sweepHeightVal = &BtcTestnetSweepBlockHeight
		cacheBlockHeight = BtcTestnetCacheBlockHeight
	case constant.BSC_MAINNET:
		sweepBlockKey = constant.BSC_SWEEP_BLOCK
		sweepHeightVal = &BscSweepBlockHeight
		cacheBlockHeight = BscCacheBlockHeight
	case constant.BSC_TESTNET:
		sweepBlockKey = constant.BSC_TESTNET_SWEEP_BLOCK
		sweepHeightVal = &BscTestnetSweepBlockHeight
		cacheBlockHeight = BscTestnetCacheBlockHeight
	case constant.ARBITRUM_ONE:
		sweepBlockKey = constant.ARBITRUM_ONE_SWEEP_BLOCK
		sweepHeightVal = &ArbitrumOneSweepBlockHeight
		cacheBlockHeight = ArbitrumOneCacheBlockHeight
	case constant.ARBITRUM_NOVA:
		sweepBlockKey = constant.ARBITRUM_NOVA_SWEEP_BLOCK
		sweepHeightVal = &ArbitrumNovaSweepBlockHeight
		cacheBlockHeight = ArbitrumNovaCacheBlockHeight
	case constant.ARBITRUM_SEPOLIA:
		sweepBlockKey = constant.ARBITRUM_SEPOLIA_SWEEP_BLOCK
		sweepHeightVal = &ArbitrumSepoliaSweepBlockHeight
		cacheBlockHeight = ArbitrumSepoliaCacheBlockHeight
	case constant.TRON_MAINNET:
		sweepBlockKey = constant.TRON_SWEEP_BLOCK
		sweepHeightVal = &TronSweepBlockHeight
		cacheBlockHeight = TronCacheBlockHeight
	case constant.TRON_NILE:
		sweepBlockKey = constant.TRON_NILE_SWEEP_BLOCK
		sweepHeightVal = &TronNileSweepBlockHeight
		cacheBlockHeight = TronNileCacheBlockHeight
	case constant.LTC_MAINNET:
		sweepBlockKey = constant.LTC_SWEEP_BLOCK
		sweepHeightVal = &LtcSweepBlockHeight
		cacheBlockHeight = LtcCacheBlockHeight
	case constant.LTC_TESTNET:
		sweepBlockKey = constant.LTC_TESTNET_SWEEP_BLOCK
		sweepHeightVal = &LtcTestnetSweepBlockHeight
		cacheBlockHeight = LtcTestnetCacheBlockHeight
	case constant.OP_MAINNET:
		sweepBlockKey = constant.OP_SWEEP_BLOCK
		sweepHeightVal = &OpSweepBlockHeight
		cacheBlockHeight = OpCacheBlockHeight
	case constant.OP_SEPOLIA:
		sweepBlockKey = constant.OP_SEPOLIA_SWEEP_BLOCK
		sweepHeightVal = &OpSepoliaSweepBlockHeight
		cacheBlockHeight = OpSepoliaCacheBlockHeight
	case constant.SOL_MAINNET:
		sweepBlockKey = constant.SOL_SWEEP_BLOCK
		sweepHeightVal = &SolSweepBlockHeight
		cacheBlockHeight = SolCacheBlockHeight
	case constant.SOL_DEVNET:
		sweepBlockKey = constant.SOL_DEVNET_SWEEP_BLOCK
		sweepHeightVal = &SolDevnetSweepBlockHeight
		cacheBlockHeight = SolDevnetCacheBlockHeight
	case constant.TON_MAINNET:
		sweepBlockKey = constant.TON_SWEEP_BLOCK
		sweepHeightVal = &TonSweepBlockHeight
		cacheBlockHeight = TonCacheBlockHeight
	case constant.TON_TESTNET:
		sweepBlockKey = constant.TON_TESTNET_SWEEP_BLOCK
		sweepHeightVal = &TonTestnetSweepBlockHeight
		cacheBlockHeight = TonTestnetCacheBlockHeight
	case constant.XRP_MAINNET:
		sweepBlockKey = constant.XRP_SWEEP_BLOCK
		sweepHeightVal = &XrpSweepBlockHeight
		cacheBlockHeight = XrpCacheBlockHeight
	case constant.XRP_TESTNET:
		sweepBlockKey = constant.XRP_TESTNET_SWEEP_BLOCK
		sweepHeightVal = &XrpTestnetSweepBlockHeight
		cacheBlockHeight = XrpTestnetCacheBlockHeight
	case constant.BCH_MAINNET:
		sweepBlockKey = constant.BCH_SWEEP_BLOCK
		sweepHeightVal = &BchSweepBlockHeight
		cacheBlockHeight = BchCacheBlockHeight
	case constant.BCH_TESTNET:
		sweepBlockKey = constant.BCH_TESTNET_SWEEP_BLOCK
		sweepHeightVal = &BchTestnetSweepBlockHeight
		cacheBlockHeight = BchTestnetCacheBlockHeight
	case constant.POL_MAINNET:
		sweepBlockKey = constant.POL_SWEEP_BLOCK
		sweepHeightVal = &PolSweepBlockHeight
		cacheBlockHeight = PolCacheBlockHeight
	case constant.POL_TESTNET:
		sweepBlockKey = constant.POL_TESTNET_SWEEP_BLOCK
		sweepHeightVal = &PolTestnetSweepBlockHeight
		cacheBlockHeight = PolTestnetCacheBlockHeight
	case constant.AVAX_MAINNET:
		sweepBlockKey = constant.AVAX_SWEEP_BLOCK
		sweepHeightVal = &AvaxSweepBlockHeight
		cacheBlockHeight = AvaxCacheBlockHeight
	case constant.AVAX_TESTNET:
		sweepBlockKey = constant.AVAX_TESTNET_SWEEP_BLOCK
		sweepHeightVal = &AvaxTestnetSweepBlockHeight
		cacheBlockHeight = AvaxTestnetCacheBlockHeight
	case constant.BASE_MAINNET:
		sweepBlockKey = constant.BASE_SWEEP_BLOCK
		sweepHeightVal = &BaseSweepBlockHeight
		cacheBlockHeight = BaseCacheBlockHeight
	case constant.BASE_SEPOLIA:
		sweepBlockKey = constant.BASE_SEPOLIA_SWEEP_BLOCK
		sweepHeightVal = &BaseSepoliaSweepBlockHeight
		cacheBlockHeight = BaseSepoliaCacheBlockHeight
	default:
		return
	}

	sweepBlockHeightString, err := global.NODE_REDIS.Get(ctx, sweepBlockKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			*sweepHeightVal = cacheBlockHeight
		} else {
			global.NODE_LOG.Error(err.Error())
			return
		}
	} else {
		*sweepHeightVal, err = strconv.ParseInt(sweepBlockHeightString, 10, 64)
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			return
		}
	}
}

func UpdatePublicKey(ctx context.Context, chainId uint) {
	if !utils.IsChainJoinSweep(chainId) {
		return
	}

	var err error
	var publicKeyString string
	var publicKeys *[]string

	switch chainId {
	case constant.ETH_MAINNET:
		publicKeyString = constant.ETH_PUBLIC_KEY
		publicKeys = &EthPublicKey
	case constant.ETH_SEPOLIA:
		publicKeyString = constant.ETH_SEPOLIA_PUBLIC_KEY
		publicKeys = &EthSepoliaPublicKey
	case constant.BTC_MAINNET:
		publicKeyString = constant.BTC_PUBLIC_KEY
		publicKeys = &BtcPublicKey
	case constant.BTC_TESTNET:
		publicKeyString = constant.BTC_TESTNET_PUBLIC_KEY
		publicKeys = &BtcTestnetPublicKey
	case constant.BSC_MAINNET:
		publicKeyString = constant.BSC_PUBLIC_KEY
		publicKeys = &BscPublicKey
	case constant.BSC_TESTNET:
		publicKeyString = constant.BSC_TESTNET_PUBLIC_KEY
		publicKeys = &BscTestnetPublicKey
	case constant.ARBITRUM_ONE:
		publicKeyString = constant.ARBITRUM_ONE_PUBLIC_KEY
		publicKeys = &ArbitrumOnePublicKey
	case constant.ARBITRUM_NOVA:
		publicKeyString = constant.ARBITRUM_NOVA_PUBLIC_KEY
		publicKeys = &ArbitrumNovaPublicKey
	case constant.ARBITRUM_SEPOLIA:
		publicKeyString = constant.ARBITRUM_SEPOLIA_PUBLIC_KEY
		publicKeys = &ArbitrumSepoliaPublicKey
	case constant.TRON_MAINNET:
		publicKeyString = constant.TRON_PUBLIC_KEY
		publicKeys = &TronPublicKey
	case constant.TRON_NILE:
		publicKeyString = constant.TRON_NILE_PUBLIC_KEY
		publicKeys = &TronNilePublicKey
	case constant.LTC_MAINNET:
		publicKeyString = constant.LTC_PUBLIC_KEY
		publicKeys = &LtcPublicKey
	case constant.LTC_TESTNET:
		publicKeyString = constant.LTC_TESTNET_PUBLIC_KEY
		publicKeys = &LtcTestnetPublicKey
	case constant.OP_MAINNET:
		publicKeyString = constant.OP_PUBLIC_KEY
		publicKeys = &OpPublicKey
	case constant.OP_SEPOLIA:
		publicKeyString = constant.OP_SEPOLIA_PUBLIC_KEY
		publicKeys = &OpSepoliaPublicKey
	case constant.SOL_MAINNET:
		publicKeyString = constant.SOL_PUBLIC_KEY
		publicKeys = &SolPublicKey
	case constant.SOL_DEVNET:
		publicKeyString = constant.SOL_DEVNET_PUBLIC_KEY
		publicKeys = &SolDevnetPublicKey
	case constant.TON_MAINNET:
		publicKeyString = constant.TON_PUBLIC_KEY
		publicKeys = &TonPublicKey
	case constant.TON_TESTNET:
		publicKeyString = constant.TON_TESTNET_PUBLIC_KEY
		publicKeys = &TonTestnetPublicKey
	case constant.XRP_MAINNET:
		publicKeyString = constant.XRP_PUBLIC_KEY
		publicKeys = &XrpPublicKey
	case constant.XRP_TESTNET:
		publicKeyString = constant.XRP_TESTNET_PUBLIC_KEY
		publicKeys = &XrpTestnetPublicKey
	case constant.BCH_MAINNET:
		publicKeyString = constant.BCH_PUBLIC_KEY
		publicKeys = &BchPublicKey
	case constant.BCH_TESTNET:
		publicKeyString = constant.BCH_TESTNET_PUBLIC_KEY
		publicKeys = &BchTestnetPublicKey
	case constant.POL_MAINNET:
		publicKeyString = constant.POL_PUBLIC_KEY
		publicKeys = &PolPublicKey
	case constant.POL_TESTNET:
		publicKeyString = constant.POL_TESTNET_PUBLIC_KEY
		publicKeys = &PolTestnetPublicKey
	case constant.AVAX_MAINNET:
		publicKeyString = constant.AVAX_PUBLIC_KEY
		publicKeys = &AvaxPublicKey
	case constant.AVAX_TESTNET:
		publicKeyString = constant.AVAX_TESTNET_PUBLIC_KEY
		publicKeys = &AvaxTestnetPublicKey
	case constant.BASE_MAINNET:
		publicKeyString = constant.BASE_PUBLIC_KEY
		publicKeys = &BasePublicKey
	case constant.BASE_SEPOLIA:
		publicKeyString = constant.BASE_SEPOLIA_PUBLIC_KEY
		publicKeys = &BaseSepoliaPublicKey
	default:
		return
	}

	pLen, err := global.NODE_REDIS.LLen(ctx, publicKeyString).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return
		}
		global.NODE_LOG.Error(err.Error())
		UpdatePublicKey(ctx, chainId)
		return
	}

	if pLen > 0 {
		*publicKeys = []string{}
		var p int64 = 0
		for ; p < pLen; p++ {
			key, err := global.NODE_REDIS.LIndex(ctx, publicKeyString, p).Result()
			if err != nil {
				global.NODE_LOG.Error(err.Error())

				p -= 1
				continue
			}
			*publicKeys = append(*publicKeys, key)
		}
	}
}

func UpdateCacheBlockHeight(ctx context.Context, chainId uint) {
	if !utils.IsChainJoinSweep(chainId) {
		return
	}

	var err error
	var cacheBlockString string
	var latestBlockHeight *int64

	switch chainId {
	case constant.ETH_MAINNET:
		cacheBlockString = constant.ETH_CACHE_BLOCK
		latestBlockHeight = &EthLatestBlockHeight
		EthCacheBlockHeight = EthLatestBlockHeight
	case constant.ETH_SEPOLIA:
		cacheBlockString = constant.ETH_SEPOLIA_CACHE_BLOCK
		latestBlockHeight = &EthSepoliaLatestBlockHeight
		EthSepoliaCacheBlockHeight = EthSepoliaLatestBlockHeight
	case constant.BSC_MAINNET:
		cacheBlockString = constant.BSC_CACHE_BLOCK
		latestBlockHeight = &BscLatestBlockHeight
		BscCacheBlockHeight = BscLatestBlockHeight
	case constant.BSC_TESTNET:
		cacheBlockString = constant.BSC_TESTNET_CACHE_BLOCK
		latestBlockHeight = &BscTestnetLatestBlockHeight
		BscTestnetCacheBlockHeight = BscTestnetLatestBlockHeight
	case constant.BTC_TESTNET:
		cacheBlockString = constant.BTC_TESTNET_CACHE_BLOCK
		latestBlockHeight = &BtcTestnetLatestBlockHeight
		BtcTestnetCacheBlockHeight = BtcTestnetLatestBlockHeight
	case constant.BTC_MAINNET:
		cacheBlockString = constant.BTC_CACHE_BLOCK
		latestBlockHeight = &BtcLatestBlockHeight
		BtcCacheBlockHeight = BtcLatestBlockHeight
	case constant.TRON_MAINNET:
		cacheBlockString = constant.TRON_CACHE_BLOCK
		latestBlockHeight = &TronLatestBlockHeight
		TronCacheBlockHeight = TronLatestBlockHeight
	case constant.TRON_NILE:
		cacheBlockString = constant.TRON_NILE_CACHE_BLOCK
		latestBlockHeight = &TronNileLatestBlockHeight
		TronNileCacheBlockHeight = TronNileLatestBlockHeight
	case constant.LTC_TESTNET:
		cacheBlockString = constant.LTC_TESTNET_CACHE_BLOCK
		latestBlockHeight = &LtcTestnetLatestBlockHeight
		LtcTestnetCacheBlockHeight = LtcTestnetLatestBlockHeight
	case constant.LTC_MAINNET:
		cacheBlockString = constant.LTC_CACHE_BLOCK
		latestBlockHeight = &LtcLatestBlockHeight
		LtcCacheBlockHeight = LtcLatestBlockHeight
	case constant.OP_MAINNET:
		cacheBlockString = constant.OP_CACHE_BLOCK
		latestBlockHeight = &OpLatestBlockHeight
		OpCacheBlockHeight = OpLatestBlockHeight
	case constant.OP_SEPOLIA:
		cacheBlockString = constant.OP_SEPOLIA_CACHE_BLOCK
		latestBlockHeight = &OpSepoliaLatestBlockHeight
		OpSepoliaCacheBlockHeight = OpSepoliaLatestBlockHeight
	case constant.ARBITRUM_ONE:
		cacheBlockString = constant.ARBITRUM_ONE_CACHE_BLOCK
		latestBlockHeight = &ArbitrumOneLatestBlockHeight
		ArbitrumOneCacheBlockHeight = ArbitrumOneLatestBlockHeight
	case constant.ARBITRUM_NOVA:
		cacheBlockString = constant.ARBITRUM_NOVA_CACHE_BLOCK
		latestBlockHeight = &ArbitrumNovaLatestBlockHeight
		ArbitrumNovaCacheBlockHeight = ArbitrumNovaLatestBlockHeight
	case constant.ARBITRUM_SEPOLIA:
		cacheBlockString = constant.ARBITRUM_SEPOLIA_CACHE_BLOCK
		latestBlockHeight = &ArbitrumSepoliaLatestBlockHeight
		ArbitrumSepoliaCacheBlockHeight = ArbitrumSepoliaLatestBlockHeight
	case constant.SOL_MAINNET:
		cacheBlockString = constant.SOL_CACHE_BLOCK
		latestBlockHeight = &SolLatestBlockHeight
		SolCacheBlockHeight = SolLatestBlockHeight
	case constant.SOL_DEVNET:
		cacheBlockString = constant.SOL_DEVNET_CACHE_BLOCK
		latestBlockHeight = &SolDevnetLatestBlockHeight
		SolDevnetCacheBlockHeight = SolDevnetLatestBlockHeight
	case constant.TON_MAINNET:
		cacheBlockString = constant.TON_CACHE_BLOCK
		latestBlockHeight = &TonLatestBlockHeight
		TonCacheBlockHeight = TonLatestBlockHeight
	case constant.TON_TESTNET:
		cacheBlockString = constant.TON_TESTNET_CACHE_BLOCK
		latestBlockHeight = &TonTestnetLatestBlockHeight
		TonTestnetCacheBlockHeight = TonTestnetLatestBlockHeight
	case constant.XRP_MAINNET:
		cacheBlockString = constant.XRP_CACHE_BLOCK
		latestBlockHeight = &XrpLatestBlockHeight
		XrpCacheBlockHeight = XrpLatestBlockHeight
	case constant.XRP_TESTNET:
		cacheBlockString = constant.XRP_TESTNET_CACHE_BLOCK
		latestBlockHeight = &XrpTestnetLatestBlockHeight
		XrpTestnetCacheBlockHeight = XrpTestnetLatestBlockHeight
	case constant.BCH_MAINNET:
		cacheBlockString = constant.BCH_CACHE_BLOCK
		latestBlockHeight = &BchLatestBlockHeight
		BchCacheBlockHeight = BchLatestBlockHeight
	case constant.BCH_TESTNET:
		cacheBlockString = constant.BCH_TESTNET_CACHE_BLOCK
		latestBlockHeight = &BchTestnetLatestBlockHeight
		BchTestnetCacheBlockHeight = BchTestnetLatestBlockHeight
	case constant.POL_MAINNET:
		cacheBlockString = constant.POL_CACHE_BLOCK
		latestBlockHeight = &PolLatestBlockHeight
		PolCacheBlockHeight = PolLatestBlockHeight
	case constant.POL_TESTNET:
		cacheBlockString = constant.POL_TESTNET_CACHE_BLOCK
		latestBlockHeight = &PolTestnetLatestBlockHeight
		PolTestnetCacheBlockHeight = PolTestnetLatestBlockHeight
	case constant.AVAX_MAINNET:
		cacheBlockString = constant.AVAX_CACHE_BLOCK
		latestBlockHeight = &AvaxLatestBlockHeight
		AvaxCacheBlockHeight = AvaxLatestBlockHeight
	case constant.AVAX_TESTNET:
		cacheBlockString = constant.AVAX_TESTNET_CACHE_BLOCK
		latestBlockHeight = &AvaxTestnetLatestBlockHeight
		AvaxTestnetCacheBlockHeight = AvaxTestnetLatestBlockHeight
	case constant.BASE_MAINNET:
		cacheBlockString = constant.BASE_CACHE_BLOCK
		latestBlockHeight = &BaseLatestBlockHeight
		BaseCacheBlockHeight = BaseLatestBlockHeight
	case constant.BASE_SEPOLIA:
		cacheBlockString = constant.BASE_SEPOLIA_CACHE_BLOCK
		latestBlockHeight = &BaseSepoliaLatestBlockHeight
		BaseSepoliaCacheBlockHeight = BaseSepoliaLatestBlockHeight
	default:
		return
	}

	_, err = global.NODE_REDIS.Set(ctx, cacheBlockString, *latestBlockHeight, 0).Result()
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		UpdateCacheBlockHeight(ctx, chainId)
	}
}

func UpdateSweepBlockHeight(ctx context.Context, chainId uint) {
	if !utils.IsChainJoinSweep(chainId) {
		return
	}

	var err error
	var sweepBlockString string
	var cacheBlockHeight *int64

	switch chainId {
	case constant.ETH_MAINNET:
		sweepBlockString = constant.ETH_SWEEP_BLOCK
		cacheBlockHeight = &EthCacheBlockHeight
		EthSweepBlockHeight = EthCacheBlockHeight
	case constant.ETH_SEPOLIA:
		sweepBlockString = constant.ETH_SEPOLIA_SWEEP_BLOCK
		cacheBlockHeight = &EthSepoliaCacheBlockHeight
		EthSepoliaSweepBlockHeight = EthSepoliaCacheBlockHeight
	case constant.BSC_MAINNET:
		sweepBlockString = constant.BSC_SWEEP_BLOCK
		cacheBlockHeight = &BscCacheBlockHeight
		BscSweepBlockHeight = BscCacheBlockHeight
	case constant.BSC_TESTNET:
		sweepBlockString = constant.BSC_TESTNET_SWEEP_BLOCK
		cacheBlockHeight = &BscTestnetCacheBlockHeight
		BscTestnetSweepBlockHeight = BscTestnetCacheBlockHeight
	case constant.BTC_TESTNET:
		sweepBlockString = constant.BTC_TESTNET_SWEEP_BLOCK
		cacheBlockHeight = &BtcTestnetCacheBlockHeight
		BtcTestnetSweepBlockHeight = BtcTestnetCacheBlockHeight
	case constant.BTC_MAINNET:
		sweepBlockString = constant.BTC_SWEEP_BLOCK
		cacheBlockHeight = &BtcCacheBlockHeight
		BtcSweepBlockHeight = BtcCacheBlockHeight
	case constant.TRON_MAINNET:
		sweepBlockString = constant.TRON_SWEEP_BLOCK
		cacheBlockHeight = &TronCacheBlockHeight
		TronSweepBlockHeight = TronCacheBlockHeight
	case constant.TRON_NILE:
		sweepBlockString = constant.TRON_NILE_SWEEP_BLOCK
		cacheBlockHeight = &TronNileCacheBlockHeight
		TronNileSweepBlockHeight = TronNileCacheBlockHeight
	case constant.LTC_TESTNET:
		sweepBlockString = constant.LTC_TESTNET_SWEEP_BLOCK
		cacheBlockHeight = &LtcTestnetCacheBlockHeight
		LtcTestnetSweepBlockHeight = LtcTestnetCacheBlockHeight
	case constant.LTC_MAINNET:
		sweepBlockString = constant.LTC_SWEEP_BLOCK
		cacheBlockHeight = &LtcCacheBlockHeight
		LtcSweepBlockHeight = LtcCacheBlockHeight
	case constant.OP_MAINNET:
		sweepBlockString = constant.OP_SWEEP_BLOCK
		cacheBlockHeight = &OpCacheBlockHeight
		OpSweepBlockHeight = OpCacheBlockHeight
	case constant.OP_SEPOLIA:
		sweepBlockString = constant.OP_SEPOLIA_SWEEP_BLOCK
		cacheBlockHeight = &OpSepoliaCacheBlockHeight
		OpSepoliaSweepBlockHeight = OpSepoliaCacheBlockHeight
	case constant.ARBITRUM_ONE:
		sweepBlockString = constant.ARBITRUM_ONE_SWEEP_BLOCK
		cacheBlockHeight = &ArbitrumOneCacheBlockHeight
		ArbitrumOneSweepBlockHeight = ArbitrumOneCacheBlockHeight
	case constant.ARBITRUM_NOVA:
		sweepBlockString = constant.ARBITRUM_NOVA_SWEEP_BLOCK
		cacheBlockHeight = &ArbitrumNovaCacheBlockHeight
		ArbitrumNovaSweepBlockHeight = ArbitrumNovaCacheBlockHeight
	case constant.ARBITRUM_SEPOLIA:
		sweepBlockString = constant.ARBITRUM_SEPOLIA_SWEEP_BLOCK
		cacheBlockHeight = &ArbitrumSepoliaCacheBlockHeight
		ArbitrumSepoliaSweepBlockHeight = ArbitrumSepoliaCacheBlockHeight
	case constant.SOL_MAINNET:
		sweepBlockString = constant.SOL_SWEEP_BLOCK
		cacheBlockHeight = &SolCacheBlockHeight
		SolSweepBlockHeight = SolCacheBlockHeight
	case constant.SOL_DEVNET:
		sweepBlockString = constant.SOL_DEVNET_SWEEP_BLOCK
		cacheBlockHeight = &SolDevnetCacheBlockHeight
		SolDevnetSweepBlockHeight = SolDevnetCacheBlockHeight
	case constant.TON_MAINNET:
		sweepBlockString = constant.TON_SWEEP_BLOCK
		cacheBlockHeight = &TonCacheBlockHeight
		TonSweepBlockHeight = TonCacheBlockHeight
	case constant.TON_TESTNET:
		sweepBlockString = constant.TON_TESTNET_SWEEP_BLOCK
		cacheBlockHeight = &TonTestnetCacheBlockHeight
		TonTestnetSweepBlockHeight = TonTestnetCacheBlockHeight
	case constant.XRP_MAINNET:
		sweepBlockString = constant.XRP_SWEEP_BLOCK
		cacheBlockHeight = &XrpCacheBlockHeight
		XrpSweepBlockHeight = XrpCacheBlockHeight
	case constant.XRP_TESTNET:
		sweepBlockString = constant.XRP_TESTNET_SWEEP_BLOCK
		cacheBlockHeight = &XrpTestnetCacheBlockHeight
		XrpTestnetSweepBlockHeight = XrpTestnetCacheBlockHeight
	case constant.BCH_MAINNET:
		sweepBlockString = constant.BCH_SWEEP_BLOCK
		cacheBlockHeight = &BchCacheBlockHeight
		BchSweepBlockHeight = BchCacheBlockHeight
	case constant.BCH_TESTNET:
		sweepBlockString = constant.BCH_TESTNET_SWEEP_BLOCK
		cacheBlockHeight = &BchTestnetCacheBlockHeight
		BchTestnetSweepBlockHeight = BchTestnetCacheBlockHeight
	case constant.POL_MAINNET:
		sweepBlockString = constant.POL_SWEEP_BLOCK
		cacheBlockHeight = &PolCacheBlockHeight
		PolSweepBlockHeight = PolCacheBlockHeight
	case constant.POL_TESTNET:
		sweepBlockString = constant.POL_TESTNET_SWEEP_BLOCK
		cacheBlockHeight = &PolTestnetCacheBlockHeight
		PolTestnetSweepBlockHeight = PolTestnetCacheBlockHeight
	case constant.AVAX_MAINNET:
		sweepBlockString = constant.AVAX_SWEEP_BLOCK
		cacheBlockHeight = &AvaxCacheBlockHeight
		AvaxSweepBlockHeight = AvaxCacheBlockHeight
	case constant.AVAX_TESTNET:
		sweepBlockString = constant.AVAX_TESTNET_SWEEP_BLOCK
		cacheBlockHeight = &AvaxTestnetCacheBlockHeight
		AvaxTestnetSweepBlockHeight = AvaxTestnetCacheBlockHeight
	case constant.BASE_MAINNET:
		sweepBlockString = constant.BASE_SWEEP_BLOCK
		cacheBlockHeight = &BaseCacheBlockHeight
		BaseSweepBlockHeight = BaseCacheBlockHeight
	case constant.BASE_SEPOLIA:
		sweepBlockString = constant.BASE_SEPOLIA_SWEEP_BLOCK
		cacheBlockHeight = &BaseSepoliaCacheBlockHeight
		BaseSepoliaSweepBlockHeight = BaseSepoliaCacheBlockHeight
	default:
		return
	}

	_, err = global.NODE_REDIS.Set(ctx, sweepBlockString, *cacheBlockHeight, 0).Result()
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		UpdateSweepBlockHeight(ctx, chainId)
	}
}

func SavePublicKeyToRedis(ctx context.Context, chainId uint, address string) (err error) {

	if !utils.IsChainJoinSweep(chainId) {
		return errors.New("do not support network")
	}

	if !constant.IsAddressSupport(chainId, address) {
		return errors.New("do not support address")
	}

	var publicKeyString string

	switch chainId {
	case constant.ETH_MAINNET:
		publicKeyString = constant.ETH_PUBLIC_KEY
	case constant.ETH_SEPOLIA:
		publicKeyString = constant.ETH_SEPOLIA_PUBLIC_KEY
	case constant.BSC_MAINNET:
		publicKeyString = constant.BSC_PUBLIC_KEY
	case constant.BSC_TESTNET:
		publicKeyString = constant.BSC_TESTNET_PUBLIC_KEY
	case constant.BTC_TESTNET:
		publicKeyString = constant.BTC_TESTNET_PUBLIC_KEY
	case constant.BTC_MAINNET:
		publicKeyString = constant.BTC_PUBLIC_KEY
	case constant.TRON_MAINNET:
		publicKeyString = constant.TRON_PUBLIC_KEY
	case constant.TRON_NILE:
		publicKeyString = constant.TRON_NILE_PUBLIC_KEY
	case constant.LTC_TESTNET:
		publicKeyString = constant.LTC_TESTNET_PUBLIC_KEY
	case constant.LTC_MAINNET:
		publicKeyString = constant.LTC_PUBLIC_KEY
	case constant.OP_MAINNET:
		publicKeyString = constant.OP_PUBLIC_KEY
	case constant.OP_SEPOLIA:
		publicKeyString = constant.OP_SEPOLIA_PUBLIC_KEY
	case constant.ARBITRUM_ONE:
		publicKeyString = constant.ARBITRUM_ONE_PUBLIC_KEY
	case constant.ARBITRUM_NOVA:
		publicKeyString = constant.ARBITRUM_NOVA_PUBLIC_KEY
	case constant.ARBITRUM_SEPOLIA:
		publicKeyString = constant.ARBITRUM_SEPOLIA_PUBLIC_KEY
	case constant.SOL_MAINNET:
		publicKeyString = constant.SOL_PUBLIC_KEY
	case constant.SOL_DEVNET:
		publicKeyString = constant.SOL_DEVNET_PUBLIC_KEY
	case constant.TON_MAINNET:
		publicKeyString = constant.TON_PUBLIC_KEY
	case constant.TON_TESTNET:
		publicKeyString = constant.TON_TESTNET_PUBLIC_KEY
	case constant.XRP_MAINNET:
		publicKeyString = constant.XRP_PUBLIC_KEY
	case constant.XRP_TESTNET:
		publicKeyString = constant.XRP_TESTNET_PUBLIC_KEY
	case constant.BCH_MAINNET:
		publicKeyString = constant.BCH_PUBLIC_KEY
	case constant.BCH_TESTNET:
		publicKeyString = constant.BCH_TESTNET_PUBLIC_KEY
	case constant.POL_MAINNET:
		publicKeyString = constant.POL_PUBLIC_KEY
	case constant.POL_TESTNET:
		publicKeyString = constant.POL_TESTNET_PUBLIC_KEY
	case constant.AVAX_MAINNET:
		publicKeyString = constant.AVAX_PUBLIC_KEY
	case constant.AVAX_TESTNET:
		publicKeyString = constant.AVAX_TESTNET_PUBLIC_KEY
	case constant.BASE_MAINNET:
		publicKeyString = constant.BASE_PUBLIC_KEY
	case constant.BASE_SEPOLIA:
		publicKeyString = constant.BASE_SEPOLIA_PUBLIC_KEY
	default:
		return
	}

	_, err = global.NODE_REDIS.RPush(context.Background(), publicKeyString, address).Result()
	if err != nil {
		return
	}

	global.NODE_LOG.Info(fmt.Sprintf("SavePublicKeyToRedis: %s, %s", constant.GetChainName(chainId), address))

	return nil
}

func GetBlockHeight(ctx context.Context, chainId uint) (string, string, string, error) {
	if !utils.IsChainJoinSweep(chainId) {
		return "", "", "", errors.New("not support")
	}

	var latestBlockKey, cacheBlockKey, sweepBlockKey string

	switch chainId {
	case constant.ETH_MAINNET:
		latestBlockKey = constant.ETH_LATEST_BLOCK
		cacheBlockKey = constant.ETH_CACHE_BLOCK
		sweepBlockKey = constant.ETH_SWEEP_BLOCK
	case constant.ETH_SEPOLIA:
		latestBlockKey = constant.ETH_SEPOLIA_LATEST_BLOCK
		cacheBlockKey = constant.ETH_SEPOLIA_CACHE_BLOCK
		sweepBlockKey = constant.ETH_SEPOLIA_SWEEP_BLOCK
	case constant.BTC_MAINNET:
		latestBlockKey = constant.BTC_LATEST_BLOCK
		cacheBlockKey = constant.BTC_CACHE_BLOCK
		sweepBlockKey = constant.BTC_SWEEP_BLOCK
	case constant.BTC_TESTNET:
		latestBlockKey = constant.BTC_TESTNET_LATEST_BLOCK
		cacheBlockKey = constant.BTC_TESTNET_CACHE_BLOCK
		sweepBlockKey = constant.BTC_TESTNET_SWEEP_BLOCK
	case constant.BSC_MAINNET:
		latestBlockKey = constant.BSC_LATEST_BLOCK
		cacheBlockKey = constant.BSC_CACHE_BLOCK
		sweepBlockKey = constant.BSC_SWEEP_BLOCK
	case constant.BSC_TESTNET:
		latestBlockKey = constant.BSC_TESTNET_LATEST_BLOCK
		cacheBlockKey = constant.BSC_TESTNET_CACHE_BLOCK
		sweepBlockKey = constant.BSC_TESTNET_SWEEP_BLOCK
	case constant.ARBITRUM_ONE:
		latestBlockKey = constant.ARBITRUM_ONE_LATEST_BLOCK
		cacheBlockKey = constant.ARBITRUM_ONE_CACHE_BLOCK
		sweepBlockKey = constant.ARBITRUM_ONE_SWEEP_BLOCK
	case constant.ARBITRUM_NOVA:
		latestBlockKey = constant.ARBITRUM_NOVA_LATEST_BLOCK
		cacheBlockKey = constant.ARBITRUM_NOVA_CACHE_BLOCK
		sweepBlockKey = constant.ARBITRUM_NOVA_SWEEP_BLOCK
	case constant.ARBITRUM_SEPOLIA:
		latestBlockKey = constant.ARBITRUM_SEPOLIA_LATEST_BLOCK
		cacheBlockKey = constant.ARBITRUM_SEPOLIA_CACHE_BLOCK
		sweepBlockKey = constant.ARBITRUM_SEPOLIA_SWEEP_BLOCK
	case constant.TRON_MAINNET:
		latestBlockKey = constant.TRON_LATEST_BLOCK
		cacheBlockKey = constant.TRON_CACHE_BLOCK
		sweepBlockKey = constant.TRON_SWEEP_BLOCK
	case constant.TRON_NILE:
		latestBlockKey = constant.TRON_NILE_LATEST_BLOCK
		cacheBlockKey = constant.TRON_NILE_CACHE_BLOCK
		sweepBlockKey = constant.TRON_NILE_SWEEP_BLOCK
	case constant.LTC_MAINNET:
		latestBlockKey = constant.LTC_LATEST_BLOCK
		cacheBlockKey = constant.LTC_CACHE_BLOCK
		sweepBlockKey = constant.LTC_SWEEP_BLOCK
	case constant.LTC_TESTNET:
		latestBlockKey = constant.LTC_TESTNET_LATEST_BLOCK
		cacheBlockKey = constant.LTC_TESTNET_CACHE_BLOCK
		sweepBlockKey = constant.LTC_TESTNET_SWEEP_BLOCK
	case constant.OP_MAINNET:
		latestBlockKey = constant.OP_LATEST_BLOCK
		cacheBlockKey = constant.OP_CACHE_BLOCK
		sweepBlockKey = constant.OP_SWEEP_BLOCK
	case constant.OP_SEPOLIA:
		latestBlockKey = constant.OP_SEPOLIA_LATEST_BLOCK
		cacheBlockKey = constant.OP_SEPOLIA_CACHE_BLOCK
		sweepBlockKey = constant.OP_SEPOLIA_SWEEP_BLOCK
	case constant.SOL_MAINNET:
		latestBlockKey = constant.SOL_LATEST_BLOCK
		cacheBlockKey = constant.SOL_CACHE_BLOCK
		sweepBlockKey = constant.SOL_SWEEP_BLOCK
	case constant.SOL_DEVNET:
		latestBlockKey = constant.SOL_DEVNET_LATEST_BLOCK
		cacheBlockKey = constant.SOL_DEVNET_CACHE_BLOCK
		sweepBlockKey = constant.SOL_DEVNET_SWEEP_BLOCK
	case constant.TON_MAINNET:
		latestBlockKey = constant.TON_LATEST_BLOCK
		cacheBlockKey = constant.TON_CACHE_BLOCK
		sweepBlockKey = constant.TON_SWEEP_BLOCK
	case constant.TON_TESTNET:
		latestBlockKey = constant.TON_TESTNET_LATEST_BLOCK
		cacheBlockKey = constant.TON_TESTNET_CACHE_BLOCK
		sweepBlockKey = constant.TON_TESTNET_SWEEP_BLOCK
	case constant.XRP_MAINNET:
		latestBlockKey = constant.XRP_LATEST_BLOCK
		cacheBlockKey = constant.XRP_CACHE_BLOCK
		sweepBlockKey = constant.XRP_SWEEP_BLOCK
	case constant.XRP_TESTNET:
		latestBlockKey = constant.XRP_TESTNET_LATEST_BLOCK
		cacheBlockKey = constant.XRP_TESTNET_CACHE_BLOCK
		sweepBlockKey = constant.XRP_TESTNET_SWEEP_BLOCK
	case constant.BCH_MAINNET:
		latestBlockKey = constant.BCH_LATEST_BLOCK
		cacheBlockKey = constant.BCH_CACHE_BLOCK
		sweepBlockKey = constant.BCH_SWEEP_BLOCK
	case constant.BCH_TESTNET:
		latestBlockKey = constant.BCH_TESTNET_LATEST_BLOCK
		cacheBlockKey = constant.BCH_TESTNET_CACHE_BLOCK
		sweepBlockKey = constant.BCH_TESTNET_SWEEP_BLOCK
	case constant.POL_MAINNET:
		latestBlockKey = constant.POL_LATEST_BLOCK
		cacheBlockKey = constant.POL_CACHE_BLOCK
		sweepBlockKey = constant.POL_SWEEP_BLOCK
	case constant.POL_TESTNET:
		latestBlockKey = constant.POL_TESTNET_LATEST_BLOCK
		cacheBlockKey = constant.POL_TESTNET_CACHE_BLOCK
		sweepBlockKey = constant.POL_TESTNET_SWEEP_BLOCK
	case constant.AVAX_MAINNET:
		latestBlockKey = constant.AVAX_LATEST_BLOCK
		cacheBlockKey = constant.AVAX_CACHE_BLOCK
		sweepBlockKey = constant.AVAX_SWEEP_BLOCK
	case constant.AVAX_TESTNET:
		latestBlockKey = constant.AVAX_TESTNET_LATEST_BLOCK
		cacheBlockKey = constant.AVAX_TESTNET_CACHE_BLOCK
		sweepBlockKey = constant.AVAX_TESTNET_SWEEP_BLOCK
	case constant.BASE_MAINNET:
		latestBlockKey = constant.BASE_LATEST_BLOCK
		cacheBlockKey = constant.BASE_CACHE_BLOCK
		sweepBlockKey = constant.BASE_SWEEP_BLOCK
	case constant.BASE_SEPOLIA:
		latestBlockKey = constant.BASE_SEPOLIA_LATEST_BLOCK
		cacheBlockKey = constant.BASE_SEPOLIA_CACHE_BLOCK
		sweepBlockKey = constant.BASE_SEPOLIA_SWEEP_BLOCK
	default:
		return "", "", "", errors.New("not support")
	}

	latestBlockHeightString, _ := global.NODE_REDIS.Get(ctx, latestBlockKey).Result()
	cacheBlockHeightString, _ := global.NODE_REDIS.Get(ctx, cacheBlockKey).Result()
	sweepBlockHeightString, _ := global.NODE_REDIS.Get(ctx, sweepBlockKey).Result()

	return latestBlockHeightString, cacheBlockHeightString, sweepBlockHeightString, nil
}
