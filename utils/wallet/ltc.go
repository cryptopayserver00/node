package wallet

import (
	"context"
	"fmt"
	"node/global/constant"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
)

// ---------------------------------------------------------------------------
// Litecoin 链参数
//
// 直接复用 btcd 的交易/签名代码，只需要自定义网络参数，不必引入 ltcd。
// 地址、WIF 的编码都由这些参数决定。
// ---------------------------------------------------------------------------

var (
	LtcMainNetParams = chaincfg.Params{
		Name:             "ltc-mainnet",
		Net:              wire.BitcoinNet(0xdbb6c0fb),
		DefaultPort:      "9333",
		Bech32HRPSegwit:  "ltc", // ltc1q...
		PubKeyHashAddrID: 0x30,  // L...
		ScriptHashAddrID: 0x32,  // M...
		PrivateKeyID:     0xb0,  // WIF 以 T 开头
		HDPrivateKeyID:   [4]byte{0x04, 0x88, 0xad, 0xe4},
		HDPublicKeyID:    [4]byte{0x04, 0x88, 0xb2, 0x1e},
		HDCoinType:       2,
	}

	LtcTestNetParams = chaincfg.Params{
		Name:             "ltc-testnet4",
		Net:              wire.BitcoinNet(0xf1c8d2fd),
		DefaultPort:      "19335",
		Bech32HRPSegwit:  "tltc", // tltc1q...
		PubKeyHashAddrID: 0x6f,   // m... / n...
		ScriptHashAddrID: 0x3a,   // Q...
		PrivateKeyID:     0xef,   // 与 BTC 测试网相同
		HDPrivateKeyID:   [4]byte{0x04, 0x35, 0x83, 0x94},
		HDPublicKeyID:    [4]byte{0x04, 0x35, 0x87, 0xcf},
		HDCoinType:       1,
	}
)

func init() {
	// 注册后 btcutil.DecodeAddress 才能识别 ltc1 / tltc1 前缀的 bech32 地址。
	for _, p := range []*chaincfg.Params{&LtcMainNetParams, &LtcTestNetParams} {
		if err := chaincfg.Register(p); err != nil {
			panic(fmt.Sprintf("register %s params: %v", p.Name, err))
		}
	}
}

// ---------------------------------------------------------------------------
// 网络配置
// ---------------------------------------------------------------------------

const (
	ltcDustLimit      = 5460       // Litecoin 粉尘阈值（litoshi）
	ltcMaxSupplyCoins = 84_000_000 // LTC 总量，用于金额校验
	ltcDefaultFeeRate = 2.0        // 费率接口失败时的兜底 (lit/vB)
)

func getLtcNet(chainId uint) (*btcNet, error) {
	switch chainId {
	case constant.LTC_MAINNET:
		return &btcNet{
			Params:         &LtcMainNetParams,
			APIBase:        "https://litecoinspace.org/api",
			Coin:           "LTC",
			DustLimit:      ltcDustLimit,
			MaxCoins:       ltcMaxSupplyCoins,
			DefaultFeeRate: ltcDefaultFeeRate,
			FeeAPI:         feeAPIMempool,
		}, nil
	case constant.LTC_TESTNET:
		return &btcNet{
			Params:         &LtcTestNetParams,
			APIBase:        "https://litecoinspace.org/testnet/api",
			Coin:           "LTC",
			DustLimit:      ltcDustLimit,
			MaxCoins:       ltcMaxSupplyCoins,
			DefaultFeeRate: ltcDefaultFeeRate,
			FeeAPI:         feeAPIMempool,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported ltc chainId: %d", chainId)
	}
}

// BuildLtcTransfer 构造并签名一笔 P2WPKH -> 任意地址的 LTC 转账，但不广播。
// pri 为 WIF 私钥（必须是压缩公钥格式）；pub 可选，非空时会与私钥推导出的公钥比对。
func BuildLtcTransfer(ctx context.Context, chainId uint, pri, pub, toAddress string, sendVal string) (*BtcTransferResult, error) {
	net, err := getLtcNet(chainId)
	if err != nil {
		return nil, err
	}
	return buildTransfer(ctx, net, pri, pub, toAddress, sendVal)
}

func SendLtcTransfer(ctx context.Context, chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	net, err := getLtcNet(chainId)
	if err != nil {
		return "", err
	}

	res, err := buildTransfer(ctx, net, pri, pub, toAddress, sendVal)
	if err != nil {
		return "", err
	}

	txid, err := net.Broadcast(ctx, res.RawTx)
	if err != nil {
		return "", err
	}

	if txid != res.TxID {
		return "", fmt.Errorf("txid mismatch: local=%s node=%s", res.TxID, txid)
	}

	return txid, nil
}
