package wallet

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"node/global/constant"
	"sort"
	"strconv"
	"strings"
	"time"

	NODE_Client "node/utils/http"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

// feeAPI 区分费率接口的种类。
type feeAPI int

const (
	feeAPIEsplora feeAPI = iota // GET /fee-estimates        -> {"1": 87.9, "6": 12.3, ...}
	feeAPIMempool               // GET /v1/fees/recommended -> {"halfHourFee": 5, ...}
)

// btcNet 描述一条 UTXO 链（BTC / LTC 共用）。零值字段会回退到 BTC 的默认值。
type btcNet struct {
	Params         *chaincfg.Params
	APIBase        string // Esplora / mempool 兼容 API
	Coin           string // 仅用于日志和错误信息
	DustLimit      int64  // 最小输出金额（最小单位）
	MaxCoins       int64  // 总量上限（币），用于金额校验
	DefaultFeeRate float64
	FeeAPI         feeAPI
	// AcceptWIFIDs 是额外接受的 WIF 版本字节。为空时只接受 Params.PrivateKeyID。
	// 例如 LTC 主网官方是 0xb0，但不少钱包/库导出的是 BTC 风格的 0x80。
	AcceptWIFIDs []byte
}

// wifVersion 读取 WIF 的版本字节（网络标识）。
func wifVersion(wifStr string) (byte, bool) {
	_, v, err := base58.CheckDecode(wifStr)
	if err != nil {
		return 0, false
	}
	return v, true
}

// checkWIF 校验 WIF 的网络字节是否属于当前网络，不匹配时给出可读的错误。
func (n *btcNet) checkWIF(wifStr string) error {
	v, ok := wifVersion(wifStr)
	if !ok {
		return errors.New("wif: invalid base58check encoding")
	}
	allowed := append([]byte{n.Params.PrivateKeyID}, n.AcceptWIFIDs...)
	for _, id := range allowed {
		if v == id {
			return nil
		}
	}
	return fmt.Errorf("wif private key does not match the selected network %s: version byte 0x%02x, allowed %#v",
		n.Params.Name, v, allowed)
}

func (n *btcNet) dust() int64 {
	if n.DustLimit > 0 {
		return n.DustLimit
	}
	return dustLimit
}

func (n *btcNet) maxCoins() int64 {
	if n.MaxCoins > 0 {
		return n.MaxCoins
	}
	return maxBtcSupply
}

func (n *btcNet) defaultFee() float64 {
	if n.DefaultFeeRate > 0 {
		return n.DefaultFeeRate
	}
	return defaultFeeRate
}

func getBtcNet(chainId uint) (*btcNet, error) {
	switch chainId {
	case constant.BTC_MAINNET:
		return &btcNet{Params: &chaincfg.MainNetParams, APIBase: "https://blockstream.info/api", Coin: "BTC"}, nil
	case constant.BTC_TESTNET:
		return &btcNet{Params: &chaincfg.TestNet3Params, APIBase: "https://blockstream.info/testnet/api", Coin: "BTC"}, nil
	default:
		return nil, fmt.Errorf("unsupported btc chainId: %d", chainId)
	}
}

const (
	satoshiPerBtc   = 100_000_000
	maxBtcSupply    = 21_000_000
	dustLimit       = 546 // 保守的粉尘阈值（P2WPKH 实际约 294）
	minFeeRate      = 1.0 // sat/vB
	defaultFeeRate  = 5.0 // 查询失败时的兜底费率
	rbfSequence     = wire.MaxTxInSequenceNum - 2
	inputVSizeP2WPK = 68.0 // P2WPKH 输入 vsize
	txOverheadVSize = 10.5 // version + locktime + 计数 + segwit marker/flag
)

type UTXO struct {
	Txid   string `json:"txid"`
	Vout   uint32 `json:"vout"`
	Value  int64  `json:"value"`
	Status struct {
		Confirmed bool `json:"confirmed"`
	} `json:"status"`
}

type BtcTransferResult struct {
	RawTx   string  // 已签名交易 hex
	TxID    string  // 本地计算的 txid
	Fee     int64   // 手续费 (sat)
	Change  int64   // 找零 (sat)，0 表示无找零
	FeeRate float64 // sat/vB
	From    string  // 发送地址（P2WPKH）
}

// 金额换算：字符串 -> satoshi，避免 float64 精度问题

func ParseBtcToSatoshi(s string) (int64, error) {
	return parseAmount(s, maxBtcSupply)
}

// parseAmount 把十进制字符串（8 位小数）精确换算成最小单位，maxCoins 为总量上限。
func parseAmount(s string, maxCoins int64) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty amount")
	}
	whole, frac, _ := strings.Cut(s, ".")
	if whole == "" {
		whole = "0"
	}
	if len(frac) > 8 {
		return 0, fmt.Errorf("amount has more than 8 decimals: %s", s)
	}
	frac += strings.Repeat("0", 8-len(frac))
	if !isDigits(whole) || !isDigits(frac) {
		return 0, fmt.Errorf("invalid amount: %s", s)
	}
	w, err := strconv.ParseInt(whole, 10, 64)
	if err != nil || w > maxCoins {
		return 0, fmt.Errorf("amount out of range: %s", s)
	}
	f, _ := strconv.ParseInt(frac, 10, 64)
	return w*satoshiPerBtc + f, nil
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}

// 链上数据：UTXO / 费率 / 广播

// GetUTXO 返回地址下的全部 UTXO。
func (n *btcNet) GetUTXO(ctx context.Context, address string) ([]UTXO, error) {
	c := NODE_Client.Client{
		URL:     fmt.Sprintf("%s/address/%s/utxo", n.APIBase, address),
		Timeout: 20 * time.Second,
	}
	var utxos []UTXO
	if err := c.HTTPGet(ctx, &utxos); err != nil {
		return nil, fmt.Errorf("fetch utxo: %w", err)
	}
	return utxos, nil
}

// EstimateFeeRate 返回约 6 个区块确认的费率 (sat/vB)。
func (n *btcNet) EstimateFeeRate(ctx context.Context) float64 {
	c := NODE_Client.Client{
		URL:     fmt.Sprintf("%s/fee-estimates", n.APIBase),
		Timeout: 20 * time.Second,
	}
	var est map[string]float64
	if err := c.HTTPGet(ctx, &est); err != nil {
		return defaultFeeRate
	}
	for _, k := range []string{"6", "12", "3", "2", "1"} {
		if v, ok := est[k]; ok && v > 0 {
			return math.Max(v, minFeeRate)
		}
	}
	return defaultFeeRate
}

// Broadcast 广播交易，返回节点确认的 txid。
func (n *btcNet) Broadcast(ctx context.Context, rawHex string) (string, error) {
	rawHex = strings.TrimSpace(rawHex)
	if err := validateRawTxHex(rawHex); err != nil {
		return "", fmt.Errorf("invalid raw tx, refusing to broadcast: %w", err)
	}

	c := NODE_Client.Client{
		URL:     fmt.Sprintf("%s/tx", n.APIBase),
		Timeout: 20 * time.Second,
	}
	data, err := c.HTTPPostText(ctx, rawHex)
	if err != nil {
		return "", fmt.Errorf("broadcast: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// 手续费 & 选币

// estimateFee 按 P2WPKH 输入估算 vsize 并乘以费率。
// outScriptLens 为每个输出 pkScript 的长度。
func estimateFee(nIn int, outScriptLens []int, feeRate float64) int64 {
	vsize := txOverheadVSize + inputVSizeP2WPK*float64(nIn)
	for _, l := range outScriptLens {
		vsize += float64(8 + 1 + l) // value + 长度前缀 + script
	}
	return int64(math.Ceil(vsize * feeRate))
}

type selection struct {
	inputs    []UTXO
	total     int64
	fee       int64
	change    int64 // 0 表示不找零
	hasChange bool
}

func selectUTXOs(utxos []UTXO, amount int64, destLen, changeLen int, feeRate float64) (*selection, error) {
	return selectUTXOsDust(utxos, amount, destLen, changeLen, feeRate, dustLimit)
}

// selectUTXOsDust 同 selectUTXOs，但粉尘阈值由调用方指定（LTC 为 5460）。
func selectUTXOsDust(utxos []UTXO, amount int64, destLen, changeLen int, feeRate float64, dust int64) (*selection, error) {
	return selectUTXOsWith(utxos, amount, destLen, changeLen, dust, func(nIn int, outLens []int) int64 {
		return estimateFee(nIn, outLens, feeRate)
	})
}

// feeFunc 按输入个数和各输出 pkScript 长度计算手续费。不同链/脚本类型的体积不同
// （P2WPKH 输入约 68 vB，P2PKH 输入约 148 B），由调用方提供。
type feeFunc func(nIn int, outScriptLens []int) int64

// selectUTXOsWith 是选币的通用实现：只用已确认的 UTXO，金额大的优先。
func selectUTXOsWith(utxos []UTXO, amount int64, destLen, changeLen int, dust int64, feeFn feeFunc) (*selection, error) {
	// 只用已确认的 UTXO，金额大的优先。
	confirmed := make([]UTXO, 0, len(utxos))
	for _, u := range utxos {
		if u.Status.Confirmed && u.Value > 0 {
			confirmed = append(confirmed, u)
		}
	}
	sort.Slice(confirmed, func(i, j int) bool { return confirmed[i].Value > confirmed[j].Value })

	sel := &selection{}
	for _, u := range confirmed {
		sel.inputs = append(sel.inputs, u)
		sel.total += u.Value

		feeNoChange := feeFn(len(sel.inputs), []int{destLen})
		if sel.total < amount+feeNoChange {
			continue
		}

		feeWithChange := feeFn(len(sel.inputs), []int{destLen, changeLen})
		change := sel.total - amount - feeWithChange
		if change >= dust {
			sel.hasChange, sel.change, sel.fee = true, change, feeWithChange
		} else {
			// 找零是粉尘，不创建找零输出，剩余并入手续费
			sel.fee = sel.total - amount
		}
		return sel, nil
	}

	fee := feeFn(len(sel.inputs), []int{destLen, changeLen})
	return nil, fmt.Errorf("insufficient funds: have %d sat (confirmed), need about %d sat", sel.total, amount+fee)
}

// 构造 + 签名（不联网广播）

// BuildBtcTransfer 构造并签名一笔 P2WPKH -> 任意地址的转账，但不广播。
// pri 为 WIF 私钥（必须是压缩公钥格式）；pub 可选，非空时会与私钥推导出的公钥比对。
func BuildBtcTransfer(ctx context.Context, chainId uint, pri, pub, toAddress string, sendVal string) (*BtcTransferResult, error) {
	net, err := getBtcNet(chainId)
	if err != nil {
		return nil, err
	}
	return buildTransfer(ctx, net, pri, pub, toAddress, sendVal)
}

// buildTransfer 是 BTC / LTC 共用的实现，接收 *btcNet 便于测试时注入 mock 网络。
func buildTransfer(ctx context.Context, net *btcNet, pri, pub, toAddress, sendVal string) (*BtcTransferResult, error) {
	amount, err := parseAmount(sendVal, net.maxCoins())
	if err != nil {
		return nil, err
	}
	if amount < net.dust() {
		return nil, fmt.Errorf("amount %d is below dust limit %d", amount, net.dust())
	}

	// --- 私钥 / 地址 ---
	wif, err := btcutil.DecodeWIF(pri)
	if err != nil {
		return nil, fmt.Errorf("decode wif: %w", err)
	}
	if err := net.checkWIF(pri); err != nil {
		return nil, err
	}
	if !wif.CompressPubKey {
		return nil, errors.New("segwit requires a compressed public key (use a compressed WIF)")
	}

	pubKey := wif.PrivKey.PubKey()
	pubKeyBytes := pubKey.SerializeCompressed()
	if pub != "" {
		p := strings.ToLower(strings.TrimPrefix(pub, "0x"))
		if p != hex.EncodeToString(pubKeyBytes) && p != hex.EncodeToString(pubKey.SerializeUncompressed()) {
			return nil, errors.New("public key does not match private key")
		}
	}

	pubKeyHash := btcutil.Hash160(pubKeyBytes)
	fromAddr, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, net.Params)
	if err != nil {
		return nil, fmt.Errorf("derive p2wpkh address (pubkey hash len=%d, pubkey len=%d): %w",
			len(pubKeyHash), len(pubKeyBytes), err)
	}
	fromPkScript, err := txscript.PayToAddrScript(fromAddr)
	if err != nil {
		return nil, fmt.Errorf("build sender pkScript: %w", err)
	}

	toAddr, err := btcutil.DecodeAddress(toAddress, net.Params)
	if err != nil {
		return nil, fmt.Errorf("decode destination address: %w", err)
	}
	if !toAddr.IsForNet(net.Params) {
		return nil, errors.New("destination address does not match the selected network")
	}
	toPkScript, err := txscript.PayToAddrScript(toAddr)
	if err != nil {
		return nil, fmt.Errorf("build destination pkScript: %w", err)
	}

	// --- UTXO / 费率 / 选币 ---
	utxos, err := net.GetUTXO(ctx, fromAddr.EncodeAddress())
	if err != nil {
		return nil, fmt.Errorf("get utxo for %s: %w", fromAddr.EncodeAddress(), err)
	}
	feeRate := net.EstimateFeeRate(ctx)

	sel, err := selectUTXOsDust(utxos, amount, len(toPkScript), len(fromPkScript), feeRate, net.dust())
	if err != nil {
		return nil, err
	}

	// --- 构造交易 ---
	tx := wire.NewMsgTx(2)
	prevOuts := make(map[wire.OutPoint]*wire.TxOut, len(sel.inputs))

	for _, u := range sel.inputs {
		h, err := chainhash.NewHashFromStr(u.Txid)
		if err != nil {
			return nil, fmt.Errorf("invalid utxo txid %s: %w", u.Txid, err)
		}
		op := wire.NewOutPoint(h, u.Vout)
		in := wire.NewTxIn(op, nil, nil)
		in.Sequence = rbfSequence // 开启 RBF，方便之后加价
		tx.AddTxIn(in)
		prevOuts[*op] = wire.NewTxOut(u.Value, fromPkScript)
	}

	tx.AddTxOut(wire.NewTxOut(amount, toPkScript))
	if sel.hasChange {
		tx.AddTxOut(wire.NewTxOut(sel.change, fromPkScript))
	}

	// --- 签名 (BIP143 / P2WPKH) ---
	fetcher := txscript.NewMultiPrevOutFetcher(prevOuts)
	sigHashes := txscript.NewTxSigHashes(tx, fetcher)

	// BIP143 的 scriptCode: OP_DUP OP_HASH160 <pkh> OP_EQUALVERIFY OP_CHECKSIG
	scriptCode, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_DUP).AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash).
		AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG).
		Script()
	if err != nil {
		return nil, err
	}

	for i, u := range sel.inputs {
		// 返回 witness = [signature, compressedPubKey]
		witness, err := txscript.WitnessSignature(tx, sigHashes, i, u.Value, scriptCode,
			txscript.SigHashAll, wif.PrivKey, true)
		if err != nil {
			return nil, fmt.Errorf("sign input %d: %w", i, err)
		}
		tx.TxIn[i].Witness = witness
	}

	// --- 本地执行脚本引擎，确认签名有效再返回 ---
	for i, u := range sel.inputs {
		vm, err := txscript.NewEngine(fromPkScript, tx, i, txscript.StandardVerifyFlags,
			nil, sigHashes, u.Value, fetcher)
		if err != nil {
			return nil, fmt.Errorf("verify input %d (init): %w", i, err)
		}
		if err := vm.Execute(); err != nil {
			return nil, fmt.Errorf("verify input %d: %w", i, err)
		}
	}

	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return nil, err
	}

	return &BtcTransferResult{
		RawTx:   hex.EncodeToString(buf.Bytes()),
		TxID:    tx.TxHash().String(),
		Fee:     sel.fee,
		Change:  sel.change,
		FeeRate: feeRate,
		From:    fromAddr.EncodeAddress(),
	}, nil
}

// 对外入口：构造 + 签名 + 广播，返回 txid
func SendBtcTransfer(ctx context.Context, chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	res, err := BuildBtcTransfer(ctx, chainId, pri, pub, toAddress, sendVal)
	if err != nil {
		return "", err
	}

	net, err := getBtcNet(chainId)
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

// validateRawTxHex 在广播前做本地检查：必须是纯 hex、能完整反序列化、
// 至少有一个输入和一个输出、且没有多余字节。
func validateRawTxHex(rawHex string) error {
	if strings.HasPrefix(rawHex, "0x") || strings.HasPrefix(rawHex, "0X") {
		return errors.New("hex must not have a 0x prefix")
	}
	raw, err := hex.DecodeString(rawHex)
	if err != nil {
		return fmt.Errorf("not valid hex: %w", err)
	}
	r := bytes.NewReader(raw)
	var tx wire.MsgTx
	if err := tx.Deserialize(r); err != nil {
		return fmt.Errorf("cannot deserialize tx (%d bytes, starts with %x): %w", len(raw), raw[:min(len(raw), 8)], err)
	}
	if r.Len() != 0 {
		return fmt.Errorf("%d trailing bytes after tx", r.Len())
	}
	if len(tx.TxIn) == 0 {
		return errors.New("tx has no inputs")
	}
	if len(tx.TxOut) == 0 {
		return errors.New("tx has no outputs")
	}
	return nil
}
