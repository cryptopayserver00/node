package wallet

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"node/global/constant"
	"strconv"
	"strings"
	"time"

	NODE_Client "node/utils/http"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

// ---------------------------------------------------------------------------
// BCH 与 BTC/LTC 的区别
//
//  1. 没有 segwit：用传统 P2PKH，签名放在 scriptSig（<sig+hashtype> <pubkey>），没有 witness。
//  2. 签名哈希类型必须带 FORKID：SIGHASH_ALL|SIGHASH_FORKID = 0x41。
//     摘要算法与 BIP143 相同，所以直接用 btcd 的 BIP143 实现来算摘要，只是 hashType 换成 0x41。
//  3. 地址是 CashAddr（bitcoincash:q...），旧版 1.../3... 地址也可以作为收款地址。
//  4. btcd 的脚本引擎不认识 FORKID，不能用它验签，所以这里自己验证（见 verifyBchInput）。
//  5. 体积：P2PKH 输入约 148 字节，输出 34 字节，手续费公式不同。
// ---------------------------------------------------------------------------

const (
	bchDustLimit      = 546
	bchMaxSupplyCoins = 21_000_000
	bchMinFeeRate     = 1.0 // sat/B，BCH 最低中继费率
	bchDefaultFeeRate = 1.0
	bchInputSize      = 148.0 // P2PKH 压缩公钥输入
	bchTxOverhead     = 10.0  // version + 输入/输出计数 + locktime

	// SIGHASH_ALL | SIGHASH_FORKID
	bchSigHashAllForkID = txscript.SigHashType(0x41)
)

// Blockbook API v2 根路径（不带结尾的 /）。
// BCH 没有 Blockstream 那样免费的 Esplora 接口，请换成你自己的 Blockbook，
// 或 GetBlock / Alchemy 等服务（形如 https://xxx/api/v2）。
var (
	BchMainnetAPIBase = "https://bch1.trezor.io/api/v2"
	BchTestnetAPIBase = ""
)

// ---------------------------------------------------------------------------
// 网络配置
// ---------------------------------------------------------------------------

type bchNet struct {
	Params     *chaincfg.Params // 仅用于 WIF 版本字节和旧版地址（与 BTC 相同）
	CashPrefix string           // bitcoincash / bchtest
	APIBase    string
}

func getBchNet(chainId uint) (*bchNet, error) {
	switch chainId {
	case constant.BCH_MAINNET:
		if BchMainnetAPIBase == "" {
			return nil, errors.New("bch mainnet api base is not configured (BchMainnetAPIBase)")
		}
		return &bchNet{&chaincfg.MainNetParams, "bitcoincash", BchMainnetAPIBase}, nil
	case constant.BCH_TESTNET:
		if BchTestnetAPIBase == "" {
			return nil, errors.New("bch testnet api base is not configured (BchTestnetAPIBase)")
		}
		return &bchNet{&chaincfg.TestNet3Params, "bchtest", BchTestnetAPIBase}, nil
	default:
		return nil, fmt.Errorf("unsupported bch chainId: %d", chainId)
	}
}

// Blockbook 的 value 是字符串形式的最小单位。
type bbUTXO struct {
	Txid          string `json:"txid"`
	Vout          uint32 `json:"vout"`
	Value         string `json:"value"`
	Confirmations int    `json:"confirmations"`
}

// GetUTXO 返回地址下已确认的 UTXO。address 是带前缀的 CashAddr。
func (n *bchNet) GetUTXO(ctx context.Context, address string) ([]UTXO, error) {
	c := NODE_Client.Client{
		URL:     fmt.Sprintf("%s/utxo/%s?confirmed=true", n.APIBase, url.PathEscape(address)),
		Timeout: 20 * time.Second,
	}
	var raw []bbUTXO
	if err := c.HTTPGet(ctx, &raw); err != nil {
		return nil, fmt.Errorf("fetch utxo: %w", err)
	}
	utxos := make([]UTXO, 0, len(raw))
	for _, r := range raw {
		v, err := strconv.ParseInt(r.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("decode utxo value %q: %w", r.Value, err)
		}
		u := UTXO{Txid: r.Txid, Vout: r.Vout, Value: v}
		u.Status.Confirmed = r.Confirmations > 0
		utxos = append(utxos, u)
	}
	return utxos, nil
}

// EstimateFeeRate 返回 sat/B。Blockbook 的 estimatefee 返回 BCH/kB。
func (n *bchNet) EstimateFeeRate(ctx context.Context) float64 {
	c := NODE_Client.Client{
		URL:     fmt.Sprintf("%s/estimatefee/2", n.APIBase),
		Timeout: 20 * time.Second,
	}
	var r struct {
		Result json.Number `json:"result"`
	}
	if err := c.HTTPGet(ctx, &r); err != nil {
		return bchDefaultFeeRate
	}
	perKB, err := r.Result.Float64()
	if err != nil || perKB <= 0 { // -1 表示无法估算
		return bchDefaultFeeRate
	}
	// BCH/kB -> sat/B。先取整到 0.001 sat/B，避免 2.0000000000000004 这类浮点误差被 Ceil 放大成多 1 sat。
	rate := math.Round(perKB*satoshiPerBtc/1000*1000) / 1000
	return math.Max(rate, bchMinFeeRate)
}

// Broadcast 广播交易，返回节点确认的 txid。
func (n *bchNet) Broadcast(ctx context.Context, rawHex string) (string, error) {
	rawHex = strings.TrimSpace(rawHex)
	if err := validateRawTxHex(rawHex); err != nil {
		return "", fmt.Errorf("invalid raw tx, refusing to broadcast: %w", err)
	}
	c := NODE_Client.Client{
		URL: fmt.Sprintf("%s/sendtx/", n.APIBase),
	}
	var r struct {
		Result string          `json:"result"`
		Error  json.RawMessage `json:"error"`
	}
	err := c.HTTPPost(ctx, strings.NewReader(rawHex), r)
	if err != nil {
		return "", fmt.Errorf("broadcast: %w", err)
	}
	if len(r.Error) > 0 && string(r.Error) != "null" {
		return "", fmt.Errorf("broadcast: node rejected tx: %s", r.Error)
	}
	return r.Result, nil
}

// ---------------------------------------------------------------------------
// 手续费 / 地址
// ---------------------------------------------------------------------------

// estimateBchFee 按 P2PKH 输入/输出估算字节数。签名长度偶尔少 1-2 字节，所以略有高估。
func estimateBchFee(nIn int, outScriptLens []int, feeRate float64) int64 {
	size := bchTxOverhead + bchInputSize*float64(nIn)
	for _, l := range outScriptLens {
		size += float64(8 + 1 + l)
	}
	return int64(math.Ceil(size * feeRate))
}

// decodeBchDestination 解析收款地址，返回 pkScript。
// 支持 CashAddr（带或不带前缀）和旧版 1.../3... 地址；拒绝 segwit 等其他类型。
func (n *bchNet) decodeDestination(addr string) ([]byte, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return nil, errors.New("empty address")
	}

	// CashAddr 带前缀（含 ':'），或省略前缀时以 q/p 开头。旧版地址不会以 q/p 开头。
	if strings.Contains(addr, ":") || strings.HasPrefix(strings.ToLower(addr), "q") || strings.HasPrefix(strings.ToLower(addr), "p") {
		prefix, typ, hash, err := DecodeCashAddr(addr, n.CashPrefix)
		if err != nil {
			return nil, err
		}
		if prefix != n.CashPrefix {
			return nil, fmt.Errorf("cashaddr prefix %q does not match network (want %q)", prefix, n.CashPrefix)
		}
		switch typ {
		case CashAddrTypeP2PKH:
			a, err := btcutil.NewAddressPubKeyHash(hash, n.Params)
			if err != nil {
				return nil, err
			}
			return txscript.PayToAddrScript(a)
		case CashAddrTypeP2SH:
			a, err := btcutil.NewAddressScriptHashFromHash(hash, n.Params)
			if err != nil {
				return nil, err
			}
			return txscript.PayToAddrScript(a)
		}
		return nil, fmt.Errorf("unsupported cashaddr type %d", typ)
	}

	// 旧版 base58 地址
	a, err := btcutil.DecodeAddress(addr, n.Params)
	if err != nil {
		return nil, err
	}
	switch a.(type) {
	case *btcutil.AddressPubKeyHash, *btcutil.AddressScriptHash:
	default:
		return nil, fmt.Errorf("unsupported address type %T (BCH has no segwit)", a)
	}
	if !a.IsForNet(n.Params) {
		return nil, errors.New("address does not match the selected network")
	}
	return txscript.PayToAddrScript(a)
}

// ---------------------------------------------------------------------------
// 验签（btcd 脚本引擎不支持 FORKID，所以自己验）
// ---------------------------------------------------------------------------

func verifyBchInput(tx *wire.MsgTx, sigHashes *txscript.TxSigHashes, idx int, amount int64, pkScript, pubKeyHash []byte) error {
	pushes, err := txscript.PushedData(tx.TxIn[idx].SignatureScript)
	if err != nil {
		return fmt.Errorf("parse scriptSig: %w", err)
	}
	if len(pushes) != 2 {
		return fmt.Errorf("scriptSig has %d pushes, want 2", len(pushes))
	}
	sigWithType, pubBytes := pushes[0], pushes[1]
	if len(sigWithType) < 2 || txscript.SigHashType(sigWithType[len(sigWithType)-1]) != bchSigHashAllForkID {
		return errors.New("signature does not end with SIGHASH_ALL|FORKID (0x41)")
	}
	if !bytes.Equal(btcutil.Hash160(pubBytes), pubKeyHash) {
		return errors.New("scriptSig pubkey does not hash to the spent output's pubkey hash")
	}
	pub, err := btcec.ParsePubKey(pubBytes)
	if err != nil {
		return fmt.Errorf("parse pubkey: %w", err)
	}
	sig, err := ecdsa.ParseDERSignature(sigWithType[:len(sigWithType)-1])
	if err != nil {
		return fmt.Errorf("parse signature: %w", err)
	}
	digest, err := txscript.CalcWitnessSigHash(pkScript, sigHashes, bchSigHashAllForkID, tx, idx, amount)
	if err != nil {
		return err
	}
	if !sig.Verify(digest, pub) {
		return errors.New("signature does not verify")
	}
	return nil
}

// ---------------------------------------------------------------------------
// 构造 + 签名（不联网广播）
// ---------------------------------------------------------------------------

// BuildBchTransfer 构造并签名一笔 P2PKH -> 任意地址的 BCH 转账，但不广播。
// pri 为 WIF 私钥（必须是压缩公钥格式）；pub 可选，非空时会与私钥推导出的公钥比对。
// 返回的 From 是 CashAddr。
func BuildBchTransfer(ctx context.Context, chainId uint, pri, pub, toAddress string, sendVal string) (*BtcTransferResult, error) {
	net, err := getBchNet(chainId)
	if err != nil {
		return nil, err
	}
	return buildBchTransfer(ctx, net, pri, pub, toAddress, sendVal)
}

func buildBchTransfer(ctx context.Context, net *bchNet, pri, pub, toAddress, sendVal string) (*BtcTransferResult, error) {
	amount, err := parseAmount(sendVal, bchMaxSupplyCoins)
	if err != nil {
		return nil, err
	}
	if amount < bchDustLimit {
		return nil, fmt.Errorf("amount %d is below dust limit %d", amount, bchDustLimit)
	}

	// --- 私钥 / 地址 ---
	wif, err := btcutil.DecodeWIF(pri)
	if err != nil {
		return nil, fmt.Errorf("decode wif: %w", err)
	}
	if v, _ := wifVersion(pri); v != net.Params.PrivateKeyID {
		return nil, fmt.Errorf("wif private key does not match the selected network %s: version byte 0x%02x, want 0x%02x",
			net.CashPrefix, v, net.Params.PrivateKeyID)
	}
	if !wif.CompressPubKey {
		return nil, errors.New("only compressed public keys are supported (use a compressed WIF)")
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
	fromAddr, err := btcutil.NewAddressPubKeyHash(pubKeyHash, net.Params)
	if err != nil {
		return nil, fmt.Errorf("derive p2pkh address: %w", err)
	}
	fromPkScript, err := txscript.PayToAddrScript(fromAddr)
	if err != nil {
		return nil, fmt.Errorf("build sender pkScript: %w", err)
	}
	fromCash, err := EncodeCashAddr(net.CashPrefix, CashAddrTypeP2PKH, pubKeyHash)
	if err != nil {
		return nil, err
	}

	toPkScript, err := net.decodeDestination(toAddress)
	if err != nil {
		return nil, fmt.Errorf("decode destination address: %w", err)
	}

	// --- UTXO / 费率 / 选币 ---
	utxos, err := net.GetUTXO(ctx, fromCash)
	if err != nil {
		return nil, fmt.Errorf("get utxo for %s: %w", fromCash, err)
	}
	feeRate := net.EstimateFeeRate(ctx)

	sel, err := selectUTXOsWith(utxos, amount, len(toPkScript), len(fromPkScript), bchDustLimit,
		func(nIn int, outLens []int) int64 { return estimateBchFee(nIn, outLens, feeRate) })
	if err != nil {
		return nil, err
	}

	// --- 构造交易（BCH 没有 RBF，序号保持默认的 final） ---
	tx := wire.NewMsgTx(2)
	prevOuts := make(map[wire.OutPoint]*wire.TxOut, len(sel.inputs))
	for _, u := range sel.inputs {
		h, err := chainhash.NewHashFromStr(u.Txid)
		if err != nil {
			return nil, fmt.Errorf("invalid utxo txid %s: %w", u.Txid, err)
		}
		op := wire.NewOutPoint(h, u.Vout)
		tx.AddTxIn(wire.NewTxIn(op, nil, nil))
		prevOuts[*op] = wire.NewTxOut(u.Value, fromPkScript)
	}
	tx.AddTxOut(wire.NewTxOut(amount, toPkScript))
	if sel.hasChange {
		tx.AddTxOut(wire.NewTxOut(sel.change, fromPkScript))
	}

	// --- 签名：BIP143 风格摘要 + SIGHASH_ALL|FORKID，scriptSig = <sig+0x41> <pubkey> ---
	fetcher := txscript.NewMultiPrevOutFetcher(prevOuts)
	sigHashes := txscript.NewTxSigHashes(tx, fetcher)

	for i, u := range sel.inputs {
		sig, err := txscript.RawTxInWitnessSignature(tx, sigHashes, i, u.Value, fromPkScript, bchSigHashAllForkID, wif.PrivKey)
		if err != nil {
			return nil, fmt.Errorf("sign input %d: %w", i, err)
		}
		scriptSig, err := txscript.NewScriptBuilder().AddData(sig).AddData(pubKeyBytes).Script()
		if err != nil {
			return nil, fmt.Errorf("build scriptSig %d: %w", i, err)
		}
		tx.TxIn[i].SignatureScript = scriptSig
	}

	// --- 本地验签 ---
	for i, u := range sel.inputs {
		if err := verifyBchInput(tx, sigHashes, i, u.Value, fromPkScript, pubKeyHash); err != nil {
			return nil, fmt.Errorf("verify input %d: %w", i, err)
		}
	}

	var buf bytes.Buffer
	if err := tx.SerializeNoWitness(&buf); err != nil {
		return nil, err
	}

	return &BtcTransferResult{
		RawTx:   hex.EncodeToString(buf.Bytes()),
		TxID:    tx.TxHash().String(),
		Fee:     sel.fee,
		Change:  sel.change,
		FeeRate: feeRate,
		From:    fromCash,
	}, nil
}

// ---------------------------------------------------------------------------
// 对外入口：构造 + 签名 + 广播，返回 txid
// ---------------------------------------------------------------------------

func SendBchTransfer(ctx context.Context, chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	net, err := getBchNet(chainId)
	if err != nil {
		return "", err
	}

	res, err := buildBchTransfer(ctx, net, pri, pub, toAddress, sendVal)
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
