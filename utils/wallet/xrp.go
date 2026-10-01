package wallet

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"node/global/constant"
	NODE_Client "node/utils/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"golang.org/x/crypto/pbkdf2"
)

// func SendXrpTransfer(ctx context.Context, chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {
// 	return "", nil
// }

// ---------------------------------------------------------------------------
// XRP Ledger（原生 XRP 的 Payment）
//
//  - 助记词：BIP39 英文词表 -> BIP44 路径 m/44'/144'/0'/0/0（secp256k1），与 xrpl.js 的
//    Wallet.fromMnemonic、Ledger、Trust Wallet 等一致。Xaman 的 family seed (s...) 不是助记词，
//    不在此支持。
//  - 交易按 XRPL 规范二进制编码（字段按 类型码、字段码 排序），签名哈希为 SHA512Half("STX\0"+数据)。
//  - 编码与 xrpl-py 的 encode / encode_for_signing / get_hash 逐字节比对过（见测试）。
// ---------------------------------------------------------------------------

const (
	xrpDecimals        = 6
	xrpLedgerWindow    = 20        // LastLedgerSequence = 当前账本 + 20（约 1 分钟）
	xrpMaxFeeDrops     = 2_000_000 // 费用上限 2 XRP，防止节点报出异常高费用
	xrpBip44Coin       = 144
	xrpAlphabet        = "rpshnaf39wBUDNEGHJKLM4PQRST7VWXYZ2bcdeCg65jkm8oFqi1tuvAxyz"
	xrpPrefixSign      = 0x53545800 // "STX\0"
	xrpPrefixTxID      = 0x54584E00 // "TXN\0"
	xrpTxTypePayment   = 0
	xrpMaxDropsBits    = 62 // 原生金额最大 10^17 drops，需要放进 62 位
	xrpMaxDrops        = 100_000_000_000_000_000
	xrpFlagsDefault    = 0
	xrpAccountIDPrefix = 0x00
)

var XrpTestnetRPC = "https://s.altnet.rippletest.net:51234/"

func getXrpRPC(chainId uint) (string, error) {
	switch chainId {
	case constant.XRP_TESTNET:
		return XrpTestnetRPC, nil
	default:
		return "", fmt.Errorf("unsupported xrp chainId: %d", chainId)
	}
}

type XrpTransferResult struct {
	TxBlob string // 已签名交易（大写 hex）
	TxID   string // 交易哈希（大写 hex）
	Fee    int64  // drops
	From   string // 发送地址 (r...)
}

// ---------------------------------------------------------------------------
// Base58（XRP 字母表）
// ---------------------------------------------------------------------------

func xrpB58Encode(b []byte) string {
	n := new(big.Int).SetBytes(b)
	base := big.NewInt(58)
	var out []byte
	for n.Sign() > 0 {
		var m big.Int
		n.DivMod(n, base, &m)
		out = append(out, xrpAlphabet[m.Int64()])
	}
	for _, c := range b { // 前导 0 字节 -> 'r'
		if c != 0 {
			break
		}
		out = append(out, xrpAlphabet[0])
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
}

func xrpB58Decode(s string) ([]byte, error) {
	n := new(big.Int)
	base := big.NewInt(58)
	for _, c := range s {
		i := strings.IndexRune(xrpAlphabet, c)
		if i < 0 {
			return nil, fmt.Errorf("invalid character %q", c)
		}
		n.Mul(n, base)
		n.Add(n, big.NewInt(int64(i)))
	}
	out := n.Bytes()
	zeros := 0
	for zeros < len(s) && s[zeros] == xrpAlphabet[0] {
		zeros++
	}
	return append(make([]byte, zeros), out...), nil
}

func xrpChecksum(b []byte) []byte {
	h := sha256.Sum256(b)
	h = sha256.Sum256(h[:])
	return h[:4]
}

func xrpEncodeAddress(id []byte) string {
	payload := append([]byte{xrpAccountIDPrefix}, id...)
	return xrpB58Encode(append(payload, xrpChecksum(payload)...))
}

func xrpDecodeAddress(s string) ([]byte, error) {
	b, err := xrpB58Decode(strings.TrimSpace(s))
	if err != nil || len(b) != 25 || b[0] != xrpAccountIDPrefix || !bytes.Equal(xrpChecksum(b[:21]), b[21:]) {
		return nil, fmt.Errorf("invalid xrp address %q", s)
	}
	return b[1:21], nil
}

// ---------------------------------------------------------------------------
// 密钥：助记词 -> m/44'/144'/0'/0/0
// ---------------------------------------------------------------------------

func xrpKeyFromMnemonic(mnemonic string) (*btcec.PrivateKey, error) {
	words := strings.Fields(mnemonic)
	if n := len(words); n != 12 && n != 15 && n != 18 && n != 21 && n != 24 {
		return nil, fmt.Errorf("mnemonic must have 12/15/18/21/24 words, got %d", n)
	}
	// BIP39: PBKDF2-HMAC-SHA512(mnemonic, "mnemonic"+passphrase, 2048)。英文词表为 ASCII，无需 NFKD。
	seed := pbkdf2.Key([]byte(strings.Join(words, " ")), []byte("mnemonic"), 2048, 64, sha512.New)

	key, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}
	for _, idx := range []uint32{
		hdkeychain.HardenedKeyStart + 44, hdkeychain.HardenedKeyStart + xrpBip44Coin, hdkeychain.HardenedKeyStart + 0, 0, 0,
	} {
		if key, err = key.Derive(idx); err != nil {
			return nil, err
		}
	}
	return key.ECPrivKey()
}

// xrpParseKey 返回私钥、压缩公钥和 20 字节账户 ID。pub 可选：公钥 hex 或 r... 地址。
func xrpParseKey(mnemonic, pub string) (*btcec.PrivateKey, []byte, []byte, error) {
	priv, err := xrpKeyFromMnemonic(mnemonic)
	if err != nil {
		return nil, nil, nil, err
	}
	pk := priv.PubKey().SerializeCompressed()
	id := btcutil.Hash160(pk)
	if pub = strings.TrimSpace(pub); pub != "" && !matchPubKeyHex(pub, pk) && pub != xrpEncodeAddress(id) {
		return nil, nil, nil, errors.New("public key does not match mnemonic")
	}
	return priv, pk, id, nil
}

// ---------------------------------------------------------------------------
// 二进制序列化（只覆盖 Payment 用到的字段类型）
// ---------------------------------------------------------------------------

type xrpField struct {
	typ, code int
	data      []byte // 已编码的值（Blob / AccountID 已含长度前缀）
}

func xrpFieldID(typ, field int) []byte {
	switch {
	case typ < 16 && field < 16:
		return []byte{byte(typ<<4 | field)}
	case typ < 16:
		return []byte{byte(typ << 4), byte(field)}
	case field < 16:
		return []byte{byte(field), byte(typ)}
	default:
		return []byte{0, byte(typ), byte(field)}
	}
}

func xrpVL(b []byte) []byte {
	n := len(b)
	switch {
	case n <= 192:
		return append([]byte{byte(n)}, b...)
	case n <= 12480:
		n -= 193
		return append([]byte{byte(193 + n>>8), byte(n & 0xff)}, b...)
	default:
		panic("xrp: blob too large")
	}
}

func xrpU16(v uint16) []byte { b := make([]byte, 2); binary.BigEndian.PutUint16(b, v); return b }
func xrpU32(v uint32) []byte { b := make([]byte, 4); binary.BigEndian.PutUint32(b, v); return b }

// xrpDrops 编码原生 XRP 金额：8 字节，最高位 0（非 IOU），次高位 1（正数），其余为 drops。
func xrpDrops(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v|0x4000000000000000)
	return b
}

func xrpSerialize(fields []xrpField) []byte {
	sort.Slice(fields, func(i, j int) bool {
		if fields[i].typ != fields[j].typ {
			return fields[i].typ < fields[j].typ
		}
		return fields[i].code < fields[j].code
	})
	var buf bytes.Buffer
	for _, f := range fields {
		buf.Write(xrpFieldID(f.typ, f.code))
		buf.Write(f.data)
	}
	return buf.Bytes()
}

type xrpPayment struct {
	Account, Destination []byte // 20 字节账户 ID
	Drops, FeeDrops      uint64
	Sequence             uint32
	LastLedger           uint32
	DestTag              *uint32
	SigningPubKey        []byte
}

func (p xrpPayment) fields(sig []byte) []xrpField {
	f := []xrpField{
		{1, 2, xrpU16(xrpTxTypePayment)}, // TransactionType
		{2, 2, xrpU32(xrpFlagsDefault)},  // Flags
		{2, 4, xrpU32(p.Sequence)},       // Sequence
		{2, 27, xrpU32(p.LastLedger)},    // LastLedgerSequence
		{6, 1, xrpDrops(p.Drops)},        // Amount
		{6, 8, xrpDrops(p.FeeDrops)},     // Fee
		{7, 3, xrpVL(p.SigningPubKey)},   // SigningPubKey
		{8, 1, xrpVL(p.Account)},         // Account
		{8, 3, xrpVL(p.Destination)},     // Destination
	}
	if p.DestTag != nil {
		f = append(f, xrpField{2, 14, xrpU32(*p.DestTag)}) // DestinationTag
	}
	if sig != nil {
		f = append(f, xrpField{7, 4, xrpVL(sig)}) // TxnSignature
	}
	return f
}

func sha512Half(parts ...[]byte) []byte {
	h := sha512.New()
	for _, p := range parts {
		h.Write(p)
	}
	return h.Sum(nil)[:32]
}

// signingData 返回用于签名的原始数据（含 "STX\0" 前缀）。
func (p xrpPayment) signingData() []byte {
	return append(xrpU32(xrpPrefixSign), xrpSerialize(p.fields(nil))...)
}

// sign 返回已签名交易 blob 和交易哈希（均为原始字节）。
func (p xrpPayment) sign(priv *btcec.PrivateKey) (blob, txid []byte) {
	hash := sha512Half(p.signingData())
	sig := ecdsa.Sign(priv, hash).Serialize() // DER，低 S（符合 XRPL 的 fully-canonical 要求）
	blob = xrpSerialize(p.fields(sig))
	return blob, sha512Half(xrpU32(xrpPrefixTxID), blob)
}

// ---------------------------------------------------------------------------
// JSON-RPC
// ---------------------------------------------------------------------------

func xrpRPC(ctx context.Context, rpc, method string, params map[string]any, out any) error {
	var resp struct {
		Result json.RawMessage `json:"result"`
	}
	c := NODE_Client.Client{
		URL:     rpc,
		Timeout: 45 * time.Second,
	}
	if err := c.HTTPPost(ctx, map[string]any{"method": method, "params": []any{params}}, &resp); err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	var status struct {
		Status       string `json:"status"`
		Error        string `json:"error"`
		ErrorMessage string `json:"error_message"`
	}
	if err := json.Unmarshal(resp.Result, &status); err != nil {
		return fmt.Errorf("%s: bad response: %w", method, err)
	}
	if status.Status == "error" {
		return &xrpError{Method: method, Code: status.Error, Message: status.ErrorMessage}
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(resp.Result, out)
}

type xrpError struct{ Method, Code, Message string }

func (e *xrpError) Error() string { return fmt.Sprintf("%s: %s %s", e.Method, e.Code, e.Message) }

type xrpAccountData struct {
	Balance    string `json:"Balance"`
	Sequence   uint32 `json:"Sequence"`
	OwnerCount uint32 `json:"OwnerCount"`
}

// xrpAccount 返回 nil 表示账户未激活（actNotFound）。
func xrpAccount(ctx context.Context, rpc, address string) (*xrpAccountData, error) {
	var r struct {
		AccountData xrpAccountData `json:"account_data"`
	}
	err := xrpRPC(ctx, rpc, "account_info", map[string]any{"account": address, "ledger_index": "current"}, &r)
	var xe *xrpError
	if errors.As(err, &xe) && xe.Code == "actNotFound" {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r.AccountData, nil
}

func xrpUint(s string) (uint64, error) { return strconv.ParseUint(s, 10, 64) }

// ---------------------------------------------------------------------------
// 构造 + 签名
// ---------------------------------------------------------------------------

// BuildXrpTransfer 构造并签名 XRP 转账，不广播。mnemonic 为 BIP39 助记词；amount 为 XRP（十进制字符串）。
// toAddress 为经典地址 (r...)；暂不支持 X-address 和目标标签。
func BuildXrpTransfer(ctx context.Context, chainId uint, mnemonic, pub, toAddress, amount string) (*XrpTransferResult, error) {
	rpc, err := getXrpRPC(chainId)
	if err != nil {
		return nil, err
	}
	return buildXrpTransfer(ctx, rpc, mnemonic, pub, toAddress, amount)
}

func buildXrpTransfer(ctx context.Context, rpc, mnemonic, pub, toAddress, amount string) (*XrpTransferResult, error) {
	bi, err := parseDecimal(amount, xrpDecimals)
	if err != nil {
		return nil, err
	}
	if bi.Cmp(big.NewInt(xrpMaxDrops)) > 0 {
		return nil, errors.New("amount exceeds the total XRP supply")
	}
	drops := bi.Uint64()

	priv, pubKey, fromID, err := xrpParseKey(mnemonic, pub)
	if err != nil {
		return nil, err
	}
	toID, err := xrpDecodeAddress(toAddress)
	if err != nil {
		return nil, err
	}
	from := xrpEncodeAddress(fromID)
	if bytes.Equal(fromID, toID) {
		return nil, errors.New("sender and recipient are the same address")
	}

	// 储备金：先拿到当前网络的数值
	var st struct {
		State struct {
			ValidatedLedger struct {
				ReserveBase uint64 `json:"reserve_base"`
				ReserveInc  uint64 `json:"reserve_inc"`
			} `json:"validated_ledger"`
		} `json:"state"`
	}
	if err := xrpRPC(ctx, rpc, "server_state", map[string]any{}, &st); err != nil {
		return nil, err
	}
	reserveBase, reserveInc := st.State.ValidatedLedger.ReserveBase, st.State.ValidatedLedger.ReserveInc

	var fee struct {
		Drops struct {
			OpenLedgerFee string `json:"open_ledger_fee"`
			BaseFee       string `json:"base_fee"`
		} `json:"drops"`
	}
	if err := xrpRPC(ctx, rpc, "fee", map[string]any{}, &fee); err != nil {
		return nil, err
	}
	feeDrops, err := xrpUint(fee.Drops.OpenLedgerFee)
	if err != nil {
		if feeDrops, err = xrpUint(fee.Drops.BaseFee); err != nil {
			return nil, errors.New("fee: node returned no usable fee")
		}
	}
	if feeDrops > xrpMaxFeeDrops {
		return nil, fmt.Errorf("network fee %d drops exceeds safety cap %d", feeDrops, xrpMaxFeeDrops)
	}

	acc, err := xrpAccount(ctx, rpc, from)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("sender account %s is not activated on this network", from)
	}
	bal, err := xrpUint(acc.Balance)
	if err != nil {
		return nil, fmt.Errorf("bad balance %q", acc.Balance)
	}
	required := drops + feeDrops + reserveBase + uint64(acc.OwnerCount)*reserveInc
	if bal < required {
		return nil, fmt.Errorf("insufficient XRP: balance %d drops, need %d (amount %d + fee %d + reserve %d)",
			bal, required, drops, feeDrops, reserveBase+uint64(acc.OwnerCount)*reserveInc)
	}

	dest, err := xrpAccount(ctx, rpc, toAddress)
	if err != nil {
		return nil, err
	}
	if dest == nil && drops < reserveBase {
		return nil, fmt.Errorf("recipient is not activated: first payment must be at least the base reserve (%d drops)", reserveBase)
	}

	var cur struct {
		Index uint32 `json:"ledger_current_index"`
	}
	if err := xrpRPC(ctx, rpc, "ledger_current", map[string]any{}, &cur); err != nil {
		return nil, err
	}

	p := xrpPayment{
		Account: fromID, Destination: toID, Drops: drops, FeeDrops: feeDrops,
		Sequence: acc.Sequence, LastLedger: cur.Index + xrpLedgerWindow, SigningPubKey: pubKey,
	}
	blob, txid := p.sign(priv)
	return &XrpTransferResult{
		TxBlob: strings.ToUpper(hex.EncodeToString(blob)),
		TxID:   strings.ToUpper(hex.EncodeToString(txid)),
		Fee:    int64(feeDrops),
		From:   from,
	}, nil
}

// ---------------------------------------------------------------------------
// 对外入口：构造 + 签名 + 提交，返回交易哈希
//
// 返回只表示节点接受了交易（tesSUCCESS 或进入队列），不代表已被验证进账本。
// ---------------------------------------------------------------------------

func SendXrpTransfer(ctx context.Context, chainId uint, mnemonic, pub, toAddress, amount string) (hash string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	rpc, err := getXrpRPC(chainId)
	if err != nil {
		return "", err
	}
	res, err := buildXrpTransfer(ctx, rpc, mnemonic, pub, toAddress, amount)
	if err != nil {
		return "", err
	}

	var sub struct {
		EngineResult  string `json:"engine_result"`
		EngineMessage string `json:"engine_result_message"`
		TxJSON        struct {
			Hash string `json:"hash"`
		} `json:"tx_json"`
	}
	if err := xrpRPC(ctx, rpc, "submit", map[string]any{"tx_blob": res.TxBlob}, &sub); err != nil {
		return "", err
	}
	if sub.EngineResult != "tesSUCCESS" && sub.EngineResult != "terQUEUED" {
		return "", fmt.Errorf("submit rejected: %s %s", sub.EngineResult, sub.EngineMessage)
	}
	if sub.TxJSON.Hash != "" && !strings.EqualFold(sub.TxJSON.Hash, res.TxID) {
		return "", fmt.Errorf("txid mismatch: local=%s node=%s", res.TxID, sub.TxJSON.Hash)
	}
	return res.TxID, nil
}
