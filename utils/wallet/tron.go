package wallet

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"node/global/constant"
	"strings"
	"time"

	NODE_Client "node/utils/http"

	sweepUtils "node/sweep/utils"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcutil/base58"
	"golang.org/x/crypto/sha3"
)

// ---------------------------------------------------------------------------
// TRON
//
// 交易在本地用 protobuf 编码，不使用节点的 createtransaction / triggersmartcontract：
// 这样收款地址、金额、合约调用数据都由本地决定，不依赖节点返回内容的可信度。
// 字段号对照官方 tronprotocol/protocol 的 .proto，测试里有与官方 proto 生成结果逐字节比对。
// ---------------------------------------------------------------------------

const (
	trxDecimals     = 6
	tronFeeLimitSun = 100_000_000 // TRC20 转账最多消耗 100 TRX 的能量费
	tronTxTTLMs     = 60_000

	tronContractTransfer = 1  // TransferContract
	tronContractTrigger  = 31 // TriggerSmartContract

	tronTypeURLTransfer = "type.googleapis.com/protocol.TransferContract"
	tronTypeURLTrigger  = "type.googleapis.com/protocol.TriggerSmartContract"
)

var (
	TronNileAPIBase = "https://nile.trongrid.io"
	TronAPIKey      = "" // TronGrid 的 TRON-PRO-API-KEY，可选
)

func getTronAPI(chainId uint) (string, error) {
	switch chainId {
	case constant.TRON_NILE:
		return TronNileAPIBase, nil
	default:
		return "", fmt.Errorf("unsupported tron chainId: %d", chainId)
	}
}

type TronTransferResult struct {
	TxID      string // sha256(raw_data)
	RawDataHx string // raw_data 的 protobuf hex
	SignedHex string // 完整 Transaction 的 protobuf hex，用于 /wallet/broadcasthex
	From      string // 发送地址 (T...)
}

// ---------------------------------------------------------------------------
// 地址 / 密钥
// ---------------------------------------------------------------------------

func keccak256(b []byte) []byte {
	h := sha3.NewLegacyKeccak256()
	h.Write(b)
	return h.Sum(nil)
}

// tronAddrBytes 返回 21 字节地址：0x41 || keccak256(pubkey[1:])[12:]。
func tronAddrBytes(pub *btcec.PublicKey) []byte {
	h := keccak256(pub.SerializeUncompressed()[1:])
	return append([]byte{0x41}, h[12:]...)
}

func tronEncodeAddr(addr21 []byte) string { return base58.CheckEncode(addr21[1:], 0x41) }

func tronDecodeAddr(s string) ([]byte, error) {
	payload, ver, err := base58.CheckDecode(strings.TrimSpace(s))
	if err != nil || ver != 0x41 || len(payload) != 20 {
		return nil, fmt.Errorf("invalid tron address %q", s)
	}
	return append([]byte{0x41}, payload...), nil
}

func tronParseKey(pri, pub string) (*btcec.PrivateKey, []byte, error) {
	raw, err := hex.DecodeString(strings.TrimPrefix(strings.TrimSpace(pri), "0x"))
	if err != nil || len(raw) != 32 {
		return nil, nil, errors.New("tron private key must be 64 hex characters")
	}
	priv, _ := btcec.PrivKeyFromBytes(raw)
	pk := priv.PubKey()
	addr := tronAddrBytes(pk)

	if pub != "" {
		p := strings.TrimSpace(pub)
		ok := matchPubKeyHex(p, pk.SerializeCompressed(), pk.SerializeUncompressed(), pk.SerializeUncompressed()[1:]) ||
			p == tronEncodeAddr(addr)
		if !ok {
			return nil, nil, errors.New("public key does not match private key")
		}
	}
	return priv, addr, nil
}

// ---------------------------------------------------------------------------
// 最小 protobuf 编码器（只需要 varint 和 length-delimited）
// ---------------------------------------------------------------------------

func pbVarint(b []byte, v uint64) []byte {
	var tmp [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(tmp[:], v)
	return append(b, tmp[:n]...)
}

func pbBytes(b []byte, field int, v []byte) []byte {
	if len(v) == 0 {
		return b
	}
	b = pbVarint(b, uint64(field)<<3|2)
	b = pbVarint(b, uint64(len(v)))
	return append(b, v...)
}

func pbInt(b []byte, field int, v int64) []byte {
	if v == 0 {
		return b
	}
	b = pbVarint(b, uint64(field)<<3)
	return pbVarint(b, uint64(v))
}

type tronBlockRef struct {
	refBlockBytes []byte // 2 字节：区块高度的低 2 字节
	refBlockHash  []byte // 8 字节：blockID[8:16]
	timestampMs   int64  // 区块时间，用于计算过期时间
}

// tronRawData 编码 Transaction.raw。contractType / typeURL / param 为合约信息，
// feeLimit 为 0 时省略（TRX 转账不需要）。
func tronRawData(ref tronBlockRef, contractType int64, typeURL string, param []byte, feeLimit, expirationMs, timestampMs int64) []byte {
	anyMsg := pbBytes(nil, 1, []byte(typeURL))
	anyMsg = pbBytes(anyMsg, 2, param)

	contract := pbInt(nil, 1, contractType)
	contract = pbBytes(contract, 2, anyMsg)

	var raw []byte
	raw = pbBytes(raw, 1, ref.refBlockBytes)
	raw = pbBytes(raw, 4, ref.refBlockHash)
	raw = pbInt(raw, 8, expirationMs)
	raw = pbBytes(raw, 11, contract)
	raw = pbInt(raw, 14, timestampMs)
	raw = pbInt(raw, 18, feeLimit)
	return raw
}

func tronTransferParam(owner, to []byte, amount int64) []byte {
	p := pbBytes(nil, 1, owner)
	p = pbBytes(p, 2, to)
	return pbInt(p, 3, amount)
}

func tronTriggerParam(owner, contract, data []byte) []byte {
	p := pbBytes(nil, 1, owner)
	p = pbBytes(p, 2, contract)
	return pbBytes(p, 4, data)
}

// trc20TransferData = transfer(address,uint256) 的 calldata。
func trc20TransferData(to21 []byte, amount *big.Int) ([]byte, error) {
	if amount.Sign() <= 0 || amount.BitLen() > 256 {
		return nil, errors.New("invalid token amount")
	}
	data := make([]byte, 0, 68)
	data = append(data, 0xa9, 0x05, 0x9c, 0xbb)
	data = append(data, make([]byte, 12)...)
	data = append(data, to21[1:]...)
	data = append(data, bytes.Repeat([]byte{0}, 32-len(amount.Bytes()))...)
	return append(data, amount.Bytes()...), nil
}

// tronSign 返回 txid 和 65 字节签名 r||s||recid。
func tronSign(priv *btcec.PrivateKey, raw []byte) (txid [32]byte, sig []byte) {
	txid = sha256.Sum256(raw)
	compact := ecdsa.SignCompact(priv, txid[:], false) // [27+recid | r | s]
	sig = append(append([]byte{}, compact[1:]...), compact[0]-27)
	return
}

// tronSignedTx 编码完整 Transaction{raw_data=1, signature=2}。
func tronSignedTx(raw, sig []byte) []byte {
	b := pbBytes(nil, 1, raw)
	return pbBytes(b, 2, sig)
}

// ---------------------------------------------------------------------------
// 节点交互
// ---------------------------------------------------------------------------
func tronBlockRefNow(ctx context.Context, base string) (tronBlockRef, error) {
	c := NODE_Client.Client{
		URL: fmt.Sprintf("%s%s", base, "/wallet/getnowblock"),
		Headers: map[string]string{
			"TRON-PRO-API-KEY": "",
		},
	}
	var blk struct {
		BlockID     string `json:"blockID"`
		BlockHeader struct {
			RawData struct {
				Number    int64 `json:"number"`
				Timestamp int64 `json:"timestamp"`
			} `json:"raw_data"`
		} `json:"block_header"`
	}
	if err := c.HTTPPost(ctx, map[string]any{}, &blk); err != nil {
		return tronBlockRef{}, fmt.Errorf("getnowblock: %w", err)
	}
	id, err := hex.DecodeString(blk.BlockID)
	if err != nil || len(id) != 32 {
		return tronBlockRef{}, fmt.Errorf("getnowblock: bad blockID %q", blk.BlockID)
	}
	var num [8]byte
	binary.BigEndian.PutUint64(num[:], uint64(blk.BlockHeader.RawData.Number))
	return tronBlockRef{
		refBlockBytes: num[6:8],
		refBlockHash:  id[8:16],
		timestampMs:   blk.BlockHeader.RawData.Timestamp,
	}, nil
}

// tronTRXBalance 返回 sun。未激活账户返回 0。
func tronTRXBalance(ctx context.Context, base, addr string) (int64, error) {
	c := NODE_Client.Client{
		URL: fmt.Sprintf("%s%s", base, "/wallet/getaccount"),
		Headers: map[string]string{
			"TRON-PRO-API-KEY": "",
		},
	}
	var acc struct {
		Balance int64 `json:"balance"`
	}
	if err := c.HTTPPost(ctx, map[string]any{"address": addr, "visible": true}, &acc); err != nil {
		return 0, fmt.Errorf("getaccount: %w", err)
	}
	return acc.Balance, nil
}

func tronTokenBalance(ctx context.Context, base string, owner21 []byte, contractAddr string) (*big.Int, error) {
	c := NODE_Client.Client{
		URL: fmt.Sprintf("%s%s", base, "/wallet/triggerconstantcontract"),
		Headers: map[string]string{
			"TRON-PRO-API-KEY": "",
		},
	}
	param := make([]byte, 12, 32)
	param = append(param, owner21[1:]...)
	var res struct {
		Result struct {
			Result  bool   `json:"result"`
			Message string `json:"message"`
		} `json:"result"`
		ConstantResult []string `json:"constant_result"`
	}
	if err := c.HTTPPost(ctx, map[string]any{
		"owner_address":     tronEncodeAddr(owner21),
		"contract_address":  contractAddr,
		"function_selector": "balanceOf(address)",
		"parameter":         hex.EncodeToString(param),
		"visible":           true,
	}, &res); err != nil {
		return nil, fmt.Errorf("balanceOf: %w", err)
	}
	if !res.Result.Result || len(res.ConstantResult) == 0 {
		return nil, fmt.Errorf("balanceOf failed: %s", res.Result.Message)
	}
	b, err := hex.DecodeString(res.ConstantResult[0])
	if err != nil {
		return nil, fmt.Errorf("balanceOf: bad result: %w", err)
	}
	return new(big.Int).SetBytes(b), nil
}

func tronBroadcast(ctx context.Context, base string, signed []byte) (string, error) {
	c := NODE_Client.Client{
		URL: fmt.Sprintf("%s%s", base, "/wallet/broadcasthex"),
		Headers: map[string]string{
			"TRON-PRO-API-KEY": "",
		},
	}
	var res struct {
		Result  bool   `json:"result"`
		Code    string `json:"code"`
		TxID    string `json:"txid"`
		Message string `json:"message"`
	}
	if err := c.HTTPPost(ctx, map[string]any{"transaction": hex.EncodeToString(signed)}, &res); err != nil {
		return "", fmt.Errorf("broadcast: %w", err)
	}
	if !res.Result {
		msg := res.Message
		if b, err := hex.DecodeString(msg); err == nil { // 节点把错误信息 hex 编码了
			msg = string(b)
		}
		return "", fmt.Errorf("broadcast rejected: %s %s", res.Code, msg)
	}
	return res.TxID, nil
}

// ---------------------------------------------------------------------------
// 构造 + 签名
// ---------------------------------------------------------------------------

func buildTron(ctx context.Context, base string, priv *btcec.PrivateKey, from21 []byte, contractType int64, typeURL string, param []byte, feeLimit int64) (*TronTransferResult, error) {
	ref, err := tronBlockRefNow(ctx, base)
	if err != nil {
		return nil, err
	}
	nowMs := nowFn().UnixMilli()
	raw := tronRawData(ref, contractType, typeURL, param, feeLimit, ref.timestampMs+tronTxTTLMs, nowMs)

	txid, sig := tronSign(priv, raw)
	return &TronTransferResult{
		TxID:      hex.EncodeToString(txid[:]),
		RawDataHx: hex.EncodeToString(raw),
		SignedHex: hex.EncodeToString(tronSignedTx(raw, sig)),
		From:      tronEncodeAddr(from21),
	}, nil
}

// BuildTrxTransfer 构造并签名 TRX 转账，不广播。amount 为 TRX（十进制字符串）。
func BuildTrxTransfer(ctx context.Context, chainId uint, pri, pub, toAddress, amount string) (*TronTransferResult, error) {
	base, err := getTronAPI(chainId)
	if err != nil {
		return nil, err
	}
	return buildTrxTransfer(ctx, base, pri, pub, toAddress, amount)
}

func buildTrxTransfer(ctx context.Context, base, pri, pub, toAddress, amount string) (*TronTransferResult, error) {
	sun, err := parseDecimal(amount, trxDecimals)
	if err != nil {
		return nil, err
	}
	if !sun.IsInt64() {
		return nil, errors.New("amount too large")
	}
	priv, from21, err := tronParseKey(pri, pub)
	if err != nil {
		return nil, err
	}
	to21, err := tronDecodeAddr(toAddress)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(from21, to21) {
		return nil, errors.New("sender and recipient are the same address")
	}

	bal, err := tronTRXBalance(ctx, base, tronEncodeAddr(from21))
	if err != nil {
		return nil, err
	}
	if bal < sun.Int64() {
		return nil, fmt.Errorf("insufficient TRX: have %d sun, need %d sun (plus bandwidth/activation fees)", bal, sun.Int64())
	}

	return buildTron(ctx, base, priv, from21, tronContractTransfer, tronTypeURLTransfer,
		tronTransferParam(from21, to21, sun.Int64()), 0)
}

// BuildTronTokenTransfer 构造并签名 TRC20 转账，不广播。coin 需先用 RegisterToken 注册。
func BuildTronTokenTransfer(ctx context.Context, chainId uint, pri, pub, toAddress, coin, amount string) (*TronTransferResult, error) {
	base, err := getTronAPI(chainId)
	if err != nil {
		return nil, err
	}
	isSupport, _, contractAddress, decimals := sweepUtils.GetContractInfoByChainIdAndSymbol(chainId, coin)
	if !isSupport {
		return nil, errors.New("contract address not found")
	}
	units, err := parseDecimal(amount, decimals)
	if err != nil {
		return nil, err
	}
	priv, from21, err := tronParseKey(pri, pub)
	if err != nil {
		return nil, err
	}
	to21, err := tronDecodeAddr(toAddress)
	if err != nil {
		return nil, err
	}
	contract21, err := tronDecodeAddr(contractAddress)
	if err != nil {
		return nil, fmt.Errorf("token contract: %w", err)
	}
	if bytes.Equal(from21, to21) {
		return nil, errors.New("sender and recipient are the same address")
	}

	bal, err := tronTokenBalance(ctx, base, from21, contractAddress)
	if err != nil {
		return nil, err
	}
	if bal.Cmp(units) < 0 {
		return nil, fmt.Errorf("insufficient token balance: have %s, need %s (smallest units)", bal, units)
	}

	data, err := trc20TransferData(to21, units)
	if err != nil {
		return nil, err
	}
	return buildTron(ctx, base, priv, from21, tronContractTrigger, tronTypeURLTrigger,
		tronTriggerParam(from21, contract21, data), tronFeeLimitSun)
}

// ---------------------------------------------------------------------------
// 对外入口：构造 + 签名 + 广播，返回 txid
// ---------------------------------------------------------------------------

func SendTrxTransfer(ctx context.Context, chainId uint, pri, pub, toAddress, amount string) (hash string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	base, err := getTronAPI(chainId)
	if err != nil {
		return "", err
	}
	res, err := buildTrxTransfer(ctx, base, pri, pub, toAddress, amount)
	if err != nil {
		return "", err
	}
	return tronSend(ctx, base, res)
}

func SendTronTokenTransfer(ctx context.Context, chainId uint, pri, pub, toAddress, coin, amount string) (hash string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	base, err := getTronAPI(chainId)
	if err != nil {
		return "", err
	}
	isSupport, _, _, _ := sweepUtils.GetContractInfoByChainIdAndSymbol(chainId, coin)
	if !isSupport {
		return "", errors.New("contract address not found")
	}
	res, err := BuildTronTokenTransfer(ctx, chainId, pri, pub, toAddress, coin, amount)
	if err != nil {
		return "", err
	}
	return tronSend(ctx, base, res)
}

func tronSend(ctx context.Context, base string, res *TronTransferResult) (string, error) {
	signed, err := hex.DecodeString(res.SignedHex)
	if err != nil {
		return "", err
	}
	txid, err := tronBroadcast(ctx, base, signed)
	if err != nil {
		return "", err
	}
	if txid != "" && txid != res.TxID {
		return "", fmt.Errorf("txid mismatch: local=%s node=%s", res.TxID, txid)
	}
	return res.TxID, nil
}
