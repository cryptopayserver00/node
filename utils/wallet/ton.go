package wallet

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"node/global/constant"
	"strings"
	"sync"
	"time"

	sweepUtils "node/sweep/utils"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/jetton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

// ---------------------------------------------------------------------------
// TON（基于 github.com/xssnick/tonutils-go）
//
//  - 助记词：标准 TON 24 词助记词（不是 BIP39），由库的 wallet.FromSeed 派生密钥。
//  - 同一个助记词在不同钱包版本下地址不同。默认 V4R2，可通过 TonWalletVersion 改成
//    wallet.ConfigV5R1Final{NetworkGlobalID: wallet.TestnetGlobalID}（Tonkeeper 新版默认的 W5）等。
//    配置里的 pub 可以填公钥 hex，也可以直接填钱包地址，后者能发现版本选错。
//  - 原生 TON：普通转账；是否 bounce 跟随收款地址（UQ/0Q 不可退回，EQ/kQ 可退回；原始格式 0:hex 视为不可退回）。
//  - Jetton：向发送方的 jetton wallet 发 transfer(0xf8a7ea5)，attach 0.05 TON 作为 gas，
//    多余部分退回发送方钱包。jetton wallet 地址由 master 合约的 get_wallet_address 查询。
//  - 连接：liteserver（ADNL），连接池按链缓存复用，请求使用 sticky context 固定到同一节点。
// ---------------------------------------------------------------------------

const (
	tonDecimals         = 9
	tonNativeFeeReserve = "0.01" // 普通转账发送前要求余额里额外留出的手续费
	tonJettonExtra      = "0.01" // jetton 转账在 gas 之外额外留出的余量
)

var (
	// 测试网 liteserver 配置。生产环境建议换成自己托管的配置文件。
	TonTestnetConfigURL = "https://ton-blockchain.github.io/testnet-global.config.json"

	// 钱包版本。默认 V4R2。
	TonWalletVersion wallet.VersionConfig = wallet.V4R2

	// jetton 转账附带的 TON（gas）。标准做法是 0.05，多余部分会退回。
	TonJettonGas = "0.05"

	// true：等交易上链并返回交易哈希（更可靠，但要等几秒到十几秒）；
	// false：发出后立即返回外部消息体的哈希（注意：这不是交易哈希，浏览器里可能搜不到）。
	TonWaitConfirmation = true

	// 单次发送的整体超时。
	TonSendTimeout = 90 * time.Second
)

type tonNet struct {
	ConfigURL string
	Testnet   bool
}

func getTonNet(chainId uint) (*tonNet, error) {
	switch chainId {
	case constant.TON_TESTNET:
		return &tonNet{ConfigURL: TonTestnetConfigURL, Testnet: true}, nil
	default:
		return nil, fmt.Errorf("unsupported ton chainId: %d", chainId)
	}
}

// ---------------------------------------------------------------------------
// 连接（可在测试里替换）
// ---------------------------------------------------------------------------

// tonAPI 是转账实际用到的最小接口，*ton.APIClient 满足它，测试里可以用替身。
type tonAPI interface {
	wallet.TonAPI
	jetton.TonApi
}

type tonConn struct {
	API    tonAPI
	Sticky func(context.Context) context.Context // 可为 nil
}

var (
	tonConnect = defaultTonConnect

	tonConnMu sync.Mutex
	tonConns  = map[uint]*tonConn{}
)

func defaultTonConnect(ctx context.Context, chainId uint, n *tonNet) (*tonConn, error) {
	tonConnMu.Lock()
	defer tonConnMu.Unlock()
	if c := tonConns[chainId]; c != nil {
		return c, nil
	}

	cfg, err := liteclient.GetConfigFromUrl(ctx, n.ConfigURL)
	if err != nil {
		return nil, fmt.Errorf("get ton config: %w", err)
	}
	pool := liteclient.NewConnectionPool()
	if err := pool.AddConnectionsFromConfig(ctx, cfg); err != nil {
		return nil, fmt.Errorf("connect to ton liteservers: %w", err)
	}
	base := ton.NewAPIClient(pool, ton.ProofCheckPolicyFast)
	base.SetTrustedBlockFromConfig(cfg) // 必须在 WithRetry 包装之前调用
	api := base.WithRetry()

	c := &tonConn{API: api, Sticky: pool.StickyContext}
	tonConns[chainId] = c
	return c, nil
}

// ---------------------------------------------------------------------------
// 地址 / 钱包
// ---------------------------------------------------------------------------

// tonParseAddr 解析友好格式 (EQ/UQ/kQ/0Q) 或原始格式 (0:hex)。raw 为 true 表示原始格式。
func tonParseAddr(s string, n *tonNet) (addr *address.Address, raw bool, err error) {
	s = strings.TrimSpace(s)
	if strings.Contains(s, ":") {
		addr, err = address.ParseRawAddr(s)
		raw = true
	} else {
		addr, err = address.ParseAddr(s)
	}
	if err != nil {
		return nil, false, fmt.Errorf("invalid ton address %q: %w", s, err)
	}
	if !n.Testnet && addr.IsTestnetOnly() {
		return nil, false, fmt.Errorf("address %q is marked testnet-only", s)
	}
	return addr, raw, nil
}

type tonSession struct {
	ctx context.Context // 已绑定 sticky 节点
	api tonAPI
	w   *wallet.Wallet
	net *tonNet
}

func tonOpen(ctx context.Context, chainId uint, mnemonic, pub string) (*tonSession, error) {
	n, err := getTonNet(chainId)
	if err != nil {
		return nil, err
	}
	words := strings.Fields(mnemonic)
	if len(words) != 24 {
		return nil, fmt.Errorf("ton mnemonic must have 24 words, got %d", len(words))
	}
	conn, err := tonConnect(ctx, chainId, n)
	if err != nil {
		return nil, err
	}
	if conn.Sticky != nil {
		ctx = conn.Sticky(ctx)
	}
	w, err := wallet.FromSeed(conn.API, words, TonWalletVersion)
	if err != nil {
		return nil, fmt.Errorf("derive wallet from mnemonic: %w", err)
	}
	if err := tonCheckPub(pub, w, n); err != nil {
		return nil, err
	}
	return &tonSession{ctx: ctx, api: conn.API, w: w, net: n}, nil
}

// tonCheckPub：pub 为空不校验；可以是公钥 hex，也可以是钱包地址（友好或原始格式）。
func tonCheckPub(pub string, w *wallet.Wallet, n *tonNet) error {
	pub = strings.TrimSpace(pub)
	if pub == "" {
		return nil
	}
	if matchPubKeyHex(pub, w.PrivateKey().Public().(ed25519.PublicKey)) {
		return nil
	}
	if a, _, err := tonParseAddr(pub, n); err == nil {
		if a.Workchain() == w.WalletAddress().Workchain() && bytes.Equal(a.Data(), w.WalletAddress().Data()) {
			return nil
		}
		return fmt.Errorf("address %s does not match the wallet derived from the mnemonic (%s); check TonWalletVersion",
			pub, w.WalletAddress().String())
	}
	return errors.New("public key does not match the mnemonic")
}

func (s *tonSession) balance() (*big.Int, error) {
	block, err := s.api.CurrentMasterchainInfo(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("get masterchain info: %w", err)
	}
	bal, err := s.w.GetBalance(s.ctx, block)
	if err != nil {
		return nil, fmt.Errorf("get wallet balance: %w", err)
	}
	return bal.Nano(), nil
}

func (s *tonSession) requireBalance(need *big.Int, what string) error {
	bal, err := s.balance()
	if err != nil {
		return err
	}
	if bal.Cmp(need) < 0 {
		return fmt.Errorf("insufficient TON in %s: have %s nanoton, need at least %s nanoton (%s)",
			s.w.WalletAddress().String(), bal, need, what)
	}
	return nil
}

func (s *tonSession) send(msg *wallet.Message) (string, error) {
	if TonWaitConfirmation {
		h, err := s.w.SendManyWaitTxHash(s.ctx, []*wallet.Message{msg})
		if err != nil {
			return "", fmt.Errorf("send and wait for transaction: %w", err)
		}
		return hex.EncodeToString(h), nil
	}
	h, err := s.w.SendManyGetInMsgHash(s.ctx, []*wallet.Message{msg})
	if err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}
	return hex.EncodeToString(h), nil
}

func tonCoins(units *big.Int, decimals int) (tlb.Coins, error) {
	c, err := tlb.FromNano(units, decimals)
	if err != nil {
		return tlb.Coins{}, fmt.Errorf("invalid amount: %w", err)
	}
	return c, nil
}

func tonMustDecimal(s string) *big.Int {
	v, err := parseDecimal(s, tonDecimals)
	if err != nil {
		panic(err)
	}
	return v
}

// ---------------------------------------------------------------------------
// 构造消息
// ---------------------------------------------------------------------------

// tonNativeMessage 构造普通 TON 转账消息。
func (s *tonSession) tonNativeMessage(toAddress, sendVal string) (*wallet.Message, *big.Int, error) {
	units, err := parseDecimal(sendVal, tonDecimals)
	if err != nil {
		return nil, nil, err
	}
	to, raw, err := tonParseAddr(toAddress, s.net)
	if err != nil {
		return nil, nil, err
	}
	if bytes.Equal(to.Data(), s.w.WalletAddress().Data()) && to.Workchain() == s.w.WalletAddress().Workchain() {
		return nil, nil, errors.New("sender and recipient are the same address")
	}
	amount, err := tonCoins(units, tonDecimals)
	if err != nil {
		return nil, nil, err
	}
	bounce := !raw && to.IsBounceable()
	msg, err := s.w.BuildTransfer(to, amount, bounce, "")
	if err != nil {
		return nil, nil, fmt.Errorf("build transfer: %w", err)
	}
	return msg, units, nil
}

// tonJettonMessage 构造 jetton 转账：发给发送方 jetton wallet 的 transfer 消息。
func (s *tonSession) tonJettonMessage(chainId uint, coin, toAddress, sendVal string) (*wallet.Message, *big.Int, error) {
	isSupportContract, _, contractAddress, decimals := sweepUtils.GetContractInfoByChainIdAndSymbol(chainId, coin)
	if !isSupportContract {
		return nil, nil, errors.New("not support")
	}

	units, err := parseDecimal(sendVal, decimals)
	if err != nil {
		return nil, nil, err
	}
	to, _, err := tonParseAddr(toAddress, s.net)
	if err != nil {
		return nil, nil, err
	}
	if bytes.Equal(to.Data(), s.w.WalletAddress().Data()) && to.Workchain() == s.w.WalletAddress().Workchain() {
		return nil, nil, errors.New("sender and recipient are the same address")
	}
	master, _, err := tonParseAddr(contractAddress, s.net)
	if err != nil {
		return nil, nil, fmt.Errorf("jetton master: %w", err)
	}

	tokenWallet, err := jetton.NewJettonMasterClient(s.api, master).GetJettonWallet(s.ctx, s.w.WalletAddress())
	if err != nil {
		return nil, nil, fmt.Errorf("get sender jetton wallet: %w", err)
	}
	have, err := tokenWallet.GetBalance(s.ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("get jetton balance: %w", err)
	}
	if have.Cmp(units) < 0 {
		return nil, nil, fmt.Errorf("insufficient jetton balance: have %s, need %s (smallest units)", have, units)
	}

	amount, err := tonCoins(units, decimals)
	if err != nil {
		return nil, nil, err
	}
	// 多余的 TON 退回发送方钱包；不要求收款方收到通知（forward amount = 0）。
	payload, err := jetton.BuildTransferPayload(to, s.w.WalletAddress(), amount, tlb.ZeroCoins, nil, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("build jetton payload: %w", err)
	}
	gas, err := tlb.FromTON(TonJettonGas)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid TonJettonGas %q: %w", TonJettonGas, err)
	}
	return wallet.SimpleMessage(tokenWallet.Address(), gas, payload), units, nil
}

// ---------------------------------------------------------------------------
// 对外入口
//
// sendVal 为十进制字符串（TON 或 jetton 的单位，不是最小单位）。
// 返回值：TonWaitConfirmation 为 true 时是交易哈希(hex)，否则是外部消息体哈希(hex)。
// ---------------------------------------------------------------------------

func SendTonTransfer(ctx context.Context, chainId uint, mnemonic, pub, toAddress string, sendVal string) (hash string, err error) {
	ctx, cancel := context.WithTimeout(ctx, TonSendTimeout)
	defer cancel()

	s, err := tonOpen(ctx, chainId, mnemonic, pub)
	if err != nil {
		return "", err
	}
	msg, units, err := s.tonNativeMessage(toAddress, sendVal)
	if err != nil {
		return "", err
	}
	need := new(big.Int).Add(units, tonMustDecimal(tonNativeFeeReserve))
	if err := s.requireBalance(need, "amount + fee reserve"); err != nil {
		return "", err
	}
	return s.send(msg)
}

func SendTonTokenTransfer(ctx context.Context, chainId uint, mnemonic, pub, toAddress, coin string, sendVal string) (hash string, err error) {
	ctx, cancel := context.WithTimeout(ctx, TonSendTimeout)
	defer cancel()

	isSupportContract, _, _, _ := sweepUtils.GetContractInfoByChainIdAndSymbol(chainId, coin)
	if !isSupportContract {
		return "", errors.New("not support")
	}

	s, err := tonOpen(ctx, chainId, mnemonic, pub)
	if err != nil {
		return "", err
	}
	msg, _, err := s.tonJettonMessage(chainId, coin, toAddress, sendVal)
	if err != nil {
		return "", err
	}
	gas := msg.InternalMessage.Amount.Nano()
	need := new(big.Int).Add(gas, tonMustDecimal(tonJettonExtra))
	if err := s.requireBalance(need, "jetton gas + reserve"); err != nil {
		return "", err
	}
	return s.send(msg)
}
