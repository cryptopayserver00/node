package wallet

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"node/global/constant"
	"strings"
	"time"

	sweepUtils "node/sweep/utils"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
)

// func SendSolTransfer(ctx context.Context, chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {
// 	return "", nil
// }

// func SendSolTokenTransfer(ctx context.Context, chainId uint, pri, pub, toAddress, coin string, sendVal string) (hash string, err error) {
// 	return "", nil
// }

// ---------------------------------------------------------------------------
// Solana（基于 github.com/gagliardetto/solana-go）
//
//  - 原生 SOL：System Program Transfer。
//  - 代币：SPL Token 与 Token-2022 的 TransferChecked。收款方没有关联代币账户(ATA)时，
//    先加一条 CreateAssociatedTokenAccountIdempotent（租金由发送方支付）。
//  - 精度、token program 都以链上 mint 为准，不信任注册表里的 Decimals。
//
// 库里的 ATA Create 指令和 TransferChecked 构造器都固定使用旧版 Token 程序，
// 为了同时支持 Token-2022，这里用库构造指令后，替换 program id 重新封装。
// ---------------------------------------------------------------------------

const (
	solDecimals         = 9
	solSignatureFee     = 5000      // 每个签名 5000 lamports
	solRentExemptEmpty  = 890_880   // 0 字节账户的免租最低余额
	solATARentSPL       = 2_039_280 // 165 字节的 SPL 代币账户
	solATARentToken2022 = 2_074_080 // 带 ImmutableOwner 扩展的 Token-2022 账户
)

var (
	SolDevnetRPC  = rpc.DevNet_RPC
	SolMainnetRPC = rpc.MainNetBeta_RPC
)

func getSolRPC(chainId uint) (string, error) {
	switch chainId {
	case constant.SOL_DEVNET:
		return SolDevnetRPC, nil
	case constant.SOL_MAINNET:
		return SolMainnetRPC, nil
	default:
		return "", fmt.Errorf("unsupported solana chainId: %d", chainId)
	}
}

type SolTransferResult struct {
	Signature string // base58，即交易 ID
	TxBase64  string // 已签名交易
	From      string
}

// ---------------------------------------------------------------------------
// 密钥
// ---------------------------------------------------------------------------

// solParseKey 支持：base58 的 64 字节(seed+pub)或 32 字节 seed；hex；以及 solana-keygen 的 JSON 数组。
// pub 可选：base58 公钥或 hex，用来确认私钥没有填错。
func solParseKey(pri, pub string) (solana.PrivateKey, error) {
	pri = strings.TrimSpace(pri)
	var raw []byte
	switch {
	case strings.HasPrefix(pri, "["):
		var arr []int
		if err := json.Unmarshal([]byte(pri), &arr); err != nil {
			return nil, fmt.Errorf("invalid key json: %w", err)
		}
		for _, v := range arr {
			if v < 0 || v > 255 {
				return nil, errors.New("invalid key json: byte out of range")
			}
			raw = append(raw, byte(v))
		}
	default:
		if b, err := hex.DecodeString(strings.TrimPrefix(pri, "0x")); err == nil && (len(b) == 32 || len(b) == 64) {
			raw = b
		} else if k, err := solana.PrivateKeyFromBase58(pri); err == nil {
			raw = k
		} else if p, err := solana.PublicKeyFromBase58(pri); err == nil {
			raw = p[:] // 32 字节 seed 的 base58 形式
		}
	}

	var key solana.PrivateKey
	switch len(raw) {
	case 32:
		key = solana.PrivateKey(ed25519.NewKeyFromSeed(raw))
	case 64:
		key = solana.PrivateKey(ed25519.NewKeyFromSeed(raw[:32]))
		if hex.EncodeToString(key[32:]) != hex.EncodeToString(raw[32:]) {
			return nil, errors.New("private key: embedded public key does not match the seed")
		}
	default:
		return nil, errors.New("solana private key must be 32/64 bytes (base58, hex or JSON array)")
	}

	if pub = strings.TrimSpace(pub); pub != "" {
		if pub != key.PublicKey().String() && !matchPubKeyHex(pub, key.PublicKey().Bytes()) {
			return nil, errors.New("public key does not match private key")
		}
	}
	return key, nil
}

// ---------------------------------------------------------------------------
// 指令 / 地址
// ---------------------------------------------------------------------------

// solATAFor 计算关联代币账户地址，tokenProgram 可以是 Token 或 Token-2022。
func solATAFor(owner, mint, tokenProgram solana.PublicKey) (solana.PublicKey, error) {
	addr, _, err := solana.FindProgramAddress(
		[][]byte{owner.Bytes(), tokenProgram.Bytes(), mint.Bytes()},
		solana.SPLAssociatedTokenAccountProgramID,
	)
	return addr, err
}

// solCreateATAIdempotent = CreateAssociatedTokenAccountIdempotent（指令号 1），已存在时不报错。
func solCreateATAIdempotent(payer, ata, owner, mint, tokenProgram solana.PublicKey) solana.Instruction {
	return solana.NewInstruction(
		solana.SPLAssociatedTokenAccountProgramID,
		solana.AccountMetaSlice{
			solana.Meta(payer).WRITE().SIGNER(),
			solana.Meta(ata).WRITE(),
			solana.Meta(owner),
			solana.Meta(mint),
			solana.Meta(solana.SystemProgramID),
			solana.Meta(tokenProgram),
		},
		[]byte{1},
	)
}

// solTransferChecked 用库构造 TransferChecked，再按 tokenProgram 重新封装。
func solTransferChecked(tokenProgram, src, mint, dst, owner solana.PublicKey, amount uint64, decimals uint8) (solana.Instruction, error) {
	ix, err := token.NewTransferCheckedInstruction(amount, decimals, src, mint, dst, owner, nil).ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	if tokenProgram.Equals(solana.TokenProgramID) {
		return ix, nil
	}
	data, err := ix.Data()
	if err != nil {
		return nil, err
	}
	return solana.NewInstruction(tokenProgram, ix.Accounts(), data), nil
}

// ---------------------------------------------------------------------------
// 交易
// ---------------------------------------------------------------------------

func solSignedTx(key solana.PrivateKey, blockhash solana.Hash, ixs []solana.Instruction) (*SolTransferResult, *solana.Transaction, error) {
	payer := key.PublicKey()
	tx, err := solana.NewTransaction(ixs, blockhash, solana.TransactionPayer(payer))
	if err != nil {
		return nil, nil, fmt.Errorf("build transaction: %w", err)
	}
	if _, err := tx.Sign(func(k solana.PublicKey) *solana.PrivateKey {
		if k.Equals(payer) {
			return &key
		}
		return nil
	}); err != nil {
		return nil, nil, fmt.Errorf("sign transaction: %w", err)
	}
	if err := tx.VerifySignatures(); err != nil {
		return nil, nil, fmt.Errorf("signature self-check failed: %w", err)
	}
	b64, err := tx.ToBase64()
	if err != nil {
		return nil, nil, err
	}
	return &SolTransferResult{Signature: tx.Signatures[0].String(), TxBase64: b64, From: payer.String()}, tx, nil
}

func solUint64(v *big.Int) (uint64, error) {
	if !v.IsUint64() {
		return 0, errors.New("amount too large")
	}
	return v.Uint64(), nil
}

// BuildSolTransfer 构造并签名 SOL 转账，不广播。sendVal 为 SOL（十进制字符串）。
func BuildSolTransfer(ctx context.Context, chainId uint, pri, pub, toAddress, sendVal string) (*SolTransferResult, error) {
	url, err := getSolRPC(chainId)
	if err != nil {
		return nil, err
	}
	res, _, err := buildSolTransfer(ctx, rpc.New(url), pri, pub, toAddress, sendVal)
	return res, err
}

func buildSolTransfer(ctx context.Context, client *rpc.Client, pri, pub, toAddress, sendVal string) (*SolTransferResult, *solana.Transaction, error) {
	bi, err := parseDecimal(sendVal, solDecimals)
	if err != nil {
		return nil, nil, err
	}
	lamports, err := solUint64(bi)
	if err != nil {
		return nil, nil, err
	}
	key, err := solParseKey(pri, pub)
	if err != nil {
		return nil, nil, err
	}
	from := key.PublicKey()
	to, err := solana.PublicKeyFromBase58(strings.TrimSpace(toAddress))
	if err != nil {
		return nil, nil, fmt.Errorf("invalid solana address %q: %w", toAddress, err)
	}
	if from.Equals(to) {
		return nil, nil, errors.New("sender and recipient are the same address")
	}

	bal, err := client.GetBalance(ctx, from, rpc.CommitmentConfirmed)
	if err != nil {
		return nil, nil, fmt.Errorf("get balance: %w", err)
	}
	if bal.Value < lamports+solSignatureFee {
		return nil, nil, fmt.Errorf("insufficient SOL: have %d lamports, need %d plus %d fee", bal.Value, lamports, solSignatureFee)
	}
	destBal, err := client.GetBalance(ctx, to, rpc.CommitmentConfirmed)
	if err != nil {
		return nil, nil, fmt.Errorf("get recipient balance: %w", err)
	}
	if destBal.Value == 0 && lamports < solRentExemptEmpty {
		return nil, nil, fmt.Errorf("recipient account does not exist: first transfer must be at least %d lamports (rent exemption)", solRentExemptEmpty)
	}

	bh, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, nil, fmt.Errorf("get latest blockhash: %w", err)
	}
	ix := system.NewTransferInstruction(lamports, from, to).Build()
	return solSignedTx(key, bh.Value.Blockhash, []solana.Instruction{ix})
}

// BuildSolTokenTransfer 构造并签名 SPL / Token-2022 转账，不广播。coin 需先用 RegisterToken 注册（Address 填 mint）。
func BuildSolTokenTransfer(ctx context.Context, chainId uint, pri, pub, toAddress, coin, sendVal string) (*SolTransferResult, error) {
	url, err := getSolRPC(chainId)
	if err != nil {
		return nil, err
	}
	res, _, err := buildSolTokenTransfer(ctx, chainId, rpc.New(url), pri, pub, toAddress, coin, sendVal)
	return res, err
}

func buildSolTokenTransfer(ctx context.Context, chainId uint, client *rpc.Client, pri, pub, toAddress, coin, sendVal string) (*SolTransferResult, *solana.Transaction, error) {
	isSupportContract, _, contractAddress, contractDecimals := sweepUtils.GetContractInfoByChainIdAndSymbol(chainId, coin)
	if !isSupportContract {
		return nil, nil, errors.New("not support")
	}

	key, err := solParseKey(pri, pub)
	if err != nil {
		return nil, nil, err
	}
	from := key.PublicKey()
	to, err := solana.PublicKeyFromBase58(strings.TrimSpace(toAddress))
	if err != nil {
		return nil, nil, fmt.Errorf("invalid solana address %q: %w", toAddress, err)
	}
	if from.Equals(to) {
		return nil, nil, errors.New("sender and recipient are the same address")
	}
	mint, err := solana.PublicKeyFromBase58(contractAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("token mint %q: %w", contractAddress, err)
	}

	// 以链上 mint 为准：确定所属 token program 和精度
	mintAcc, err := client.GetAccountInfo(ctx, mint)
	if errors.Is(err, rpc.ErrNotFound) {
		return nil, nil, fmt.Errorf("mint %s does not exist on this network", mint)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("get mint: %w", err)
	}
	var program solana.PublicKey
	var ataRent uint64
	switch {
	case mintAcc.Value.Owner.Equals(solana.TokenProgramID):
		program, ataRent = solana.TokenProgramID, solATARentSPL
	case mintAcc.Value.Owner.Equals(solana.Token2022ProgramID):
		program, ataRent = solana.Token2022ProgramID, solATARentToken2022
	default:
		return nil, nil, fmt.Errorf("account %s is not a token mint (owner %s)", mint, mintAcc.Value.Owner)
	}
	var m token.Mint
	if err := bin.NewBinDecoder(mintAcc.Value.Data.GetBinary()).Decode(&m); err != nil {
		return nil, nil, fmt.Errorf("decode mint: %w", err)
	}
	if !m.IsInitialized {
		return nil, nil, fmt.Errorf("mint %s is not initialized", mint)
	}
	decimals := int(m.Decimals)
	if decimals != 0 && contractDecimals != decimals {
		return nil, nil, fmt.Errorf("registered decimals %d do not match on-chain decimals %d", contractDecimals, decimals)
	}

	bi, err := parseDecimal(sendVal, decimals)
	if err != nil {
		return nil, nil, err
	}
	units, err := solUint64(bi)
	if err != nil {
		return nil, nil, err
	}

	srcATA, err := solATAFor(from, mint, program)
	if err != nil {
		return nil, nil, err
	}
	dstATA, err := solATAFor(to, mint, program)
	if err != nil {
		return nil, nil, err
	}

	tb, err := client.GetTokenAccountBalance(ctx, srcATA, rpc.CommitmentConfirmed)
	if err != nil {
		return nil, nil, fmt.Errorf("sender token account %s: %w", srcATA, err)
	}
	have, ok := new(big.Int).SetString(tb.Value.Amount, 10)
	if !ok {
		return nil, nil, fmt.Errorf("bad token balance %q", tb.Value.Amount)
	}
	if have.Cmp(bi) < 0 {
		return nil, nil, fmt.Errorf("insufficient token balance: have %s, need %s (smallest units)", have, bi)
	}

	_, err = client.GetAccountInfo(ctx, dstATA)
	needCreate := errors.Is(err, rpc.ErrNotFound)
	if err != nil && !needCreate {
		return nil, nil, fmt.Errorf("get recipient token account: %w", err)
	}

	solBal, err := client.GetBalance(ctx, from, rpc.CommitmentConfirmed)
	if err != nil {
		return nil, nil, fmt.Errorf("get balance: %w", err)
	}
	need := uint64(solSignatureFee)
	if needCreate {
		need += ataRent
	}
	if solBal.Value < need {
		return nil, nil, fmt.Errorf("insufficient SOL for fees: have %d lamports, need about %d", solBal.Value, need)
	}

	bh, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, nil, fmt.Errorf("get latest blockhash: %w", err)
	}

	var ixs []solana.Instruction
	if needCreate {
		ixs = append(ixs, solCreateATAIdempotent(from, dstATA, to, mint, program))
	}
	xfer, err := solTransferChecked(program, srcATA, mint, dstATA, from, units, uint8(decimals))
	if err != nil {
		return nil, nil, err
	}
	ixs = append(ixs, xfer)
	return solSignedTx(key, bh.Value.Blockhash, ixs)
}

// ---------------------------------------------------------------------------
// 对外入口：构造 + 签名 + 广播，返回交易签名（即 txid）
//
// 返回只表示节点接受了交易（预检通过），不代表已经确认。
// ---------------------------------------------------------------------------

func SendSolTransfer(ctx context.Context, chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	url, err := getSolRPC(chainId)
	if err != nil {
		return "", err
	}
	client := rpc.New(url)
	_, tx, err := buildSolTransfer(ctx, client, pri, pub, toAddress, sendVal)
	if err != nil {
		return "", err
	}
	return solBroadcast(ctx, client, tx)
}

func SendSolTokenTransfer(ctx context.Context, chainId uint, pri, pub, toAddress, coin string, sendVal string) (hash string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	url, err := getSolRPC(chainId)
	if err != nil {
		return "", err
	}
	client := rpc.New(url)
	_, tx, err := buildSolTokenTransfer(ctx, chainId, client, pri, pub, toAddress, coin, sendVal)
	if err != nil {
		return "", err
	}
	return solBroadcast(ctx, client, tx)
}

func solBroadcast(ctx context.Context, client *rpc.Client, tx *solana.Transaction) (string, error) {
	sig, err := client.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
		PreflightCommitment: rpc.CommitmentConfirmed,
	})
	if err != nil {
		return "", fmt.Errorf("send transaction: %w", err)
	}
	if sig != tx.Signatures[0] {
		return "", fmt.Errorf("signature mismatch: local=%s node=%s", tx.Signatures[0], sig)
	}
	return sig.String(), nil
}
