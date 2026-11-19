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
	SweepThreshold = 10

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

	// 清空内存侧监控地址，避免重复 append
	for _, st := range chainRegistry {
		*st.PublicKeys = nil
	}

	var wallets []model.Wallet
	err = global.NODE_DB.WithContext(ctx).Select("chain_id", "address").Find(&wallets).Error
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	for _, w := range wallets {
		st, ok := GetChainState(w.ChainId)
		if !ok {
			continue
		}
		if _, err = global.NODE_REDIS.SAdd(ctx, st.PublicKeyKey, w.Address).Result(); err != nil {
			global.NODE_LOG.Error(err.Error())
			return
		}
		*st.PublicKeys = append(*st.PublicKeys, w.Address)
	}
}

func SetupLatestBlockHeight(ctx context.Context, chainId uint, blockNumber int64) {
	if blockNumber <= 0 || !utils.IsChainJoinSweep(chainId) {
		return
	}
	st, ok := GetChainState(chainId)
	if !ok {
		return
	}
	if _, err := global.NODE_REDIS.Set(ctx, st.LatestKey, blockNumber, 0).Result(); err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}
	*st.Latest = blockNumber
}

func SetupCacheBlockHeight(ctx context.Context, chainId uint) {
	if !utils.IsChainJoinSweep(chainId) {
		return
	}
	st, ok := GetChainState(chainId)
	if !ok {
		return
	}
	cacheBlockHeightString, err := global.NODE_REDIS.Get(ctx, st.CacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			*st.Cache = *st.Latest
		} else {
			global.NODE_LOG.Error(err.Error())
		}
		return
	}
	*st.Cache, err = strconv.ParseInt(cacheBlockHeightString, 10, 64)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
	}
}

func SetupSweepBlockHeight(ctx context.Context, chainId uint) {
	if !utils.IsChainJoinSweep(chainId) {
		return
	}
	st, ok := GetChainState(chainId)
	if !ok {
		return
	}
	sweepBlockHeightString, err := global.NODE_REDIS.Get(ctx, st.SweepKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			*st.Sweep = *st.Cache
		} else {
			global.NODE_LOG.Error(err.Error())
		}
		return
	}
	*st.Sweep, err = strconv.ParseInt(sweepBlockHeightString, 10, 64)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
	}
}

func UpdatePublicKey(ctx context.Context, chainId uint) {
	if !utils.IsChainJoinSweep(chainId) {
		return
	}
	st, ok := GetChainState(chainId)
	if !ok {
		return
	}
	keys, err := global.NODE_REDIS.SMembers(ctx, st.PublicKeyKey).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			global.NODE_LOG.Error(err.Error())
		}
		return
	}
	*st.PublicKeys = keys
}

func UpdateCacheBlockHeight(ctx context.Context, chainId uint) {
	if !utils.IsChainJoinSweep(chainId) {
		return
	}
	st, ok := GetChainState(chainId)
	if !ok {
		return
	}
	if *st.Latest <= 0 {
		return
	}
	*st.Cache = *st.Latest
	if _, err := global.NODE_REDIS.Set(ctx, st.CacheKey, *st.Latest, 0).Result(); err != nil {
		global.NODE_LOG.Error(err.Error())
	}
}

func UpdateSweepBlockHeight(ctx context.Context, chainId uint) {
	if !utils.IsChainJoinSweep(chainId) {
		return
	}
	st, ok := GetChainState(chainId)
	if !ok {
		return
	}
	if *st.Cache <= 0 {
		return
	}
	*st.Sweep = *st.Cache
	if _, err := global.NODE_REDIS.Set(ctx, st.SweepKey, *st.Cache, 0).Result(); err != nil {
		global.NODE_LOG.Error(err.Error())
	}
}

func SavePublicKeyToRedis(ctx context.Context, chainId uint, address string) (err error) {
	if !utils.IsChainJoinSweep(chainId) {
		return errors.New("do not support network")
	}
	if !constant.IsAddressSupport(chainId, address) {
		return errors.New("do not support address")
	}
	st, ok := GetChainState(chainId)
	if !ok {
		return errors.New("do not support network")
	}
	if _, err = global.NODE_REDIS.SAdd(ctx, st.PublicKeyKey, address).Result(); err != nil {
		return
	}
	global.NODE_LOG.Info(fmt.Sprintf("SavePublicKeyToRedis: %s, %s", constant.GetChainName(chainId), address))
	return nil
}

func GetBlockHeight(ctx context.Context, chainId uint) (string, string, string, error) {
	if !utils.IsChainJoinSweep(chainId) {
		return "", "", "", errors.New("not support")
	}
	st, ok := GetChainState(chainId)
	if !ok {
		return "", "", "", errors.New("not support")
	}
	latestBlockHeightString, _ := global.NODE_REDIS.Get(ctx, st.LatestKey).Result()
	cacheBlockHeightString, _ := global.NODE_REDIS.Get(ctx, st.CacheKey).Result()
	sweepBlockHeightString, _ := global.NODE_REDIS.Get(ctx, st.SweepKey).Result()
	return latestBlockHeightString, cacheBlockHeightString, sweepBlockHeightString, nil
}
