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
		constant.ETH_GOERLI_PUBLIC_KEY,
		constant.ETH_SEPOLIA_PUBLIC_KEY,
		constant.BTC_PUBLIC_KEY,
		constant.BTC_TESTNET_PUBLIC_KEY,
		constant.TRON_PUBLIC_KEY,
		constant.TRON_NILE_PUBLIC_KEY,
		constant.BSC_PUBLIC_KEY,
		constant.BSC_TESTNET_PUBLIC_KEY,
		constant.ARBITRUM_ONE_PUBLIC_KEY,
		constant.ARBITRUM_NOVA_PUBLIC_KEY,
		constant.ARBITRUM_GOERLI_PUBLIC_KEY,
		constant.ARBITRUM_SEPOLIA_PUBLIC_KEY,
		constant.LTC_PUBLIC_KEY,
		constant.LTC_TESTNET_PUBLIC_KEY,
		constant.OP_PUBLIC_KEY,
		constant.OP_SEPOLIA_PUBLIC_KEY,
		constant.SOL_PUBLIC_KEY,
		constant.SOL_DEVNET_PUBLIC_KEY,
		constant.TON_PUBLIC_KEY,
		constant.TON_TESTNET_PUBLIC_KEY,
	}

	EthPublicKey             []string
	BtcPublicKey             []string
	BtcTestnetPublicKey      []string
	EthGoerliPublicKey       []string
	EthSepoliaPublicKey      []string
	BscPublicKey             []string
	BscTestnetPublicKey      []string
	ArbitrumOnePublicKey     []string
	ArbitrumNovaPublicKey    []string
	ArbitrumGoerliPublicKey  []string
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

	// eth mainnet
	EthLatestBlockHeight int64
	EthCacheBlockHeight  int64
	EthSweepBlockHeight  int64

	// eth goerli
	EthGoerliLatestBlockHeight int64
	EthGoerliCacheBlockHeight  int64
	EthGoerliSweepBlockHeight  int64

	// eth sepolia
	EthSepoliaLatestBlockHeight int64
	EthSepoliaCacheBlockHeight  int64
	EthSepoliaSweepBlockHeight  int64

	// btc mainnet
	BtcLatestBlockHeight int64
	BtcCacheBlockHeight  int64
	BtcSweepBlockHeight  int64

	// btc testnet
	BtcTestnetLatestBlockHeight int64
	BtcTestnetCacheBlockHeight  int64
	BtcTestnetSweepBlockHeight  int64

	// bsc mainnet
	BscLatestBlockHeight int64
	BscCacheBlockHeight  int64
	BscSweepBlockHeight  int64

	// bsc testnet
	BscTestnetLatestBlockHeight int64
	BscTestnetCacheBlockHeight  int64
	BscTestnetSweepBlockHeight  int64

	// arbitrum one
	ArbitrumOneLatestBlockHeight int64
	ArbitrumOneCacheBlockHeight  int64
	ArbitrumOneSweepBlockHeight  int64

	// arbitrum nova
	ArbitrumNovaLatestBlockHeight int64
	ArbitrumNovaCacheBlockHeight  int64
	ArbitrumNovaSweepBlockHeight  int64

	// arbitrum goerli
	ArbitrumGoerliLatestBlockHeight int64
	ArbitrumGoerliCacheBlockHeight  int64
	ArbitrumGoerliSweepBlockHeight  int64

	// arbitrum sepolia
	ArbitrumSepoliaLatestBlockHeight int64
	ArbitrumSepoliaCacheBlockHeight  int64
	ArbitrumSepoliaSweepBlockHeight  int64

	// tron mainnet
	TronLatestBlockHeight int64
	TronCacheBlockHeight  int64
	TronSweepBlockHeight  int64

	// tron nile
	TronNileLatestBlockHeight int64
	TronNileCacheBlockHeight  int64
	TronNileSweepBlockHeight  int64

	// ltc mainnet
	LtcLatestBlockHeight int64
	LtcCacheBlockHeight  int64
	LtcSweepBlockHeight  int64

	// ltc testnet
	LtcTestnetLatestBlockHeight int64
	LtcTestnetCacheBlockHeight  int64
	LtcTestnetSweepBlockHeight  int64

	// op mainnet
	OpLatestBlockHeight int64
	OpCacheBlockHeight  int64
	OpSweepBlockHeight  int64

	// op Sepolia
	OpSepoliaLatestBlockHeight int64
	OpSepoliaCacheBlockHeight  int64
	OpSepoliaSweepBlockHeight  int64

	// sol mainnet
	SolLatestBlockHeight int64
	SolCacheBlockHeight  int64
	SolSweepBlockHeight  int64

	// sol devnet
	SolDevnetLatestBlockHeight int64
	SolDevnetCacheBlockHeight  int64
	SolDevnetSweepBlockHeight  int64

	// ton mainnet
	TonLatestBlockHeight int64
	TonCacheBlockHeight  int64
	TonSweepBlockHeight  int64

	// ton testnet
	TonTestnetLatestBlockHeight int64
	TonTestnetCacheBlockHeight  int64
	TonTestnetSweepBlockHeight  int64
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
			case constant.ETH_GOERLI:
				_, err = global.NODE_REDIS.RPush(ctx, constant.ETH_GOERLI_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				EthGoerliPublicKey = append(EthGoerliPublicKey, w.Address)
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
			case constant.ARBITRUM_GOERLI:
				_, err = global.NODE_REDIS.RPush(ctx, constant.ARBITRUM_GOERLI_PUBLIC_KEY, w.Address).Result()
				if err != nil {
					global.NODE_LOG.Error(err.Error())
					return
				}
				ArbitrumGoerliPublicKey = append(ArbitrumGoerliPublicKey, w.Address)
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
	case constant.ETH_GOERLI:
		latestBlockKey = constant.ETH_GOERLI_LATEST_BLOCK
		latestHeightVal = &EthGoerliLatestBlockHeight
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
	case constant.ARBITRUM_GOERLI:
		latestBlockKey = constant.ARBITRUM_GOERLI_LATEST_BLOCK
		latestHeightVal = &ArbitrumGoerliLatestBlockHeight
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
	case constant.ETH_GOERLI:
		cacheBlockKey = constant.ETH_GOERLI_CACHE_BLOCK
		cacheHeightVal = &EthGoerliCacheBlockHeight
		latestBlockHeight = EthGoerliLatestBlockHeight
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
	case constant.ARBITRUM_GOERLI:
		cacheBlockKey = constant.ARBITRUM_GOERLI_CACHE_BLOCK
		cacheHeightVal = &ArbitrumGoerliCacheBlockHeight
		latestBlockHeight = ArbitrumGoerliLatestBlockHeight
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
	case constant.ETH_GOERLI:
		sweepBlockKey = constant.ETH_GOERLI_SWEEP_BLOCK
		sweepHeightVal = &EthGoerliSweepBlockHeight
		cacheBlockHeight = EthGoerliCacheBlockHeight
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
	case constant.ARBITRUM_GOERLI:
		sweepBlockKey = constant.ARBITRUM_GOERLI_SWEEP_BLOCK
		sweepHeightVal = &ArbitrumGoerliSweepBlockHeight
		cacheBlockHeight = ArbitrumGoerliCacheBlockHeight
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
	case constant.ETH_GOERLI:
		publicKeyString = constant.ETH_GOERLI_PUBLIC_KEY
		publicKeys = &EthGoerliPublicKey
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
	case constant.ARBITRUM_GOERLI:
		publicKeyString = constant.ARBITRUM_GOERLI_PUBLIC_KEY
		publicKeys = &ArbitrumGoerliPublicKey
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
	case constant.ETH_GOERLI:
		cacheBlockString = constant.ETH_GOERLI_CACHE_BLOCK
		latestBlockHeight = &EthGoerliLatestBlockHeight
		EthGoerliCacheBlockHeight = EthGoerliLatestBlockHeight
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
	case constant.ARBITRUM_GOERLI:
		cacheBlockString = constant.ARBITRUM_GOERLI_CACHE_BLOCK
		latestBlockHeight = &ArbitrumGoerliLatestBlockHeight
		ArbitrumGoerliCacheBlockHeight = ArbitrumGoerliLatestBlockHeight
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
	case constant.ETH_GOERLI:
		sweepBlockString = constant.ETH_GOERLI_SWEEP_BLOCK
		cacheBlockHeight = &EthGoerliCacheBlockHeight
		EthGoerliSweepBlockHeight = EthGoerliCacheBlockHeight
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
	case constant.ARBITRUM_GOERLI:
		sweepBlockString = constant.ARBITRUM_GOERLI_SWEEP_BLOCK
		cacheBlockHeight = &ArbitrumGoerliCacheBlockHeight
		ArbitrumGoerliSweepBlockHeight = ArbitrumGoerliCacheBlockHeight
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
	case constant.ETH_GOERLI:
		publicKeyString = constant.ETH_GOERLI_PUBLIC_KEY
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
	case constant.ARBITRUM_GOERLI:
		publicKeyString = constant.ARBITRUM_GOERLI_PUBLIC_KEY
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
