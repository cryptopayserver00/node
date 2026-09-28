package wallet

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"node/global"
	"node/global/constant"
	sweepUtils "node/sweep/utils"
	"node/sweep/utils/erc20"
	"node/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	Transfer     = "transfer"
	TransferFrom = "transferFrom"
	Name         = "name"
	Symbol       = "symbol"
	Decimals     = "decimals"
	TotalSupply  = "totalSupply"
	BalanceOf    = "balanceOf"
	Approve      = "approve"

	CreateNewContract = "createNewContract"
	Withdraw          = "withdraw"

	knownMethods = map[string]string{
		"0xa9059cbb": Transfer,
		"0x23b872dd": TransferFrom,

		"0x694e974c": CreateNewContract,
	}
)

func GenerateEthereumWallet() (string, string, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return "", "", err
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	pKey := hexutil.Encode(privateKeyBytes)[2:]

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", "", errors.New("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	return pKey, address, nil
}

func CallEthTransfer(ctx context.Context, chainId uint, rpc, fromPri, fromPub, toAddress string, value *big.Int, gasLimit uint64) (hash string, err error) {
	var data []byte
	hash, err = CallWalletTransactionCore(ctx, chainId, rpc, fromPri, fromPub, toAddress, value, data, gasLimit)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func CallTokenTransfer(ctx context.Context, chainId uint, rpc, fromPri, fromPub, toAddress, tokenAddress string, tokenValue *big.Int, gasLimit uint64) (hash string, err error) {
	var value = big.NewInt(0)

	data, err := erc20.ERC20ABI.Pack(Transfer, common.HexToAddress(toAddress), tokenValue)
	if err != nil {
		return "", err
	}

	hash, err = CallWalletTransactionCore(ctx, chainId, rpc, fromPri, fromPub, tokenAddress, value, data, gasLimit)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func CallTokenName(ctx context.Context, rpc, tokenAddress string) (result any, err error) {
	result, err = CallContractCore(ctx, rpc, tokenAddress, Name)
	if err != nil {
		return nil, err
	}

	return
}

func CallTokenSymbol(ctx context.Context, rpc, tokenAddress string) (result any, err error) {
	result, err = CallContractCore(ctx, rpc, tokenAddress, Symbol)
	if err != nil {
		return nil, err
	}
	return
}

func CallTokenDecimals(ctx context.Context, rpc, tokenAddress string) (result any, err error) {
	result, err = CallContractCore(ctx, rpc, tokenAddress, Decimals)
	if err != nil {
		return nil, err
	}
	return
}

func CallTokenBalanceOf(ctx context.Context, rpc, fromPub, tokenAddress string) (balance *big.Int, err error) {
	result, err := CallContractCore(ctx, rpc, tokenAddress, BalanceOf, common.HexToAddress(fromPub))
	if err != nil {
		return nil, err
	}

	return result["balance"].(*big.Int), nil
}

func GetTransactionSenderFromTx(tx *types.Transaction) (string, error) {
	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	return from.String(), err
}

func SendEthTransfer(ctx context.Context, chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {
	rpc := constant.GetRPCUrlByNetwork(chainId)
	if rpc == "" {
		err = errors.New("chain not support")
		return
	}

	var (
		gasLimit uint64 = 21000
		decimals int    = 18
	)

	sendValue, err := utils.FormatToOriginalValue(sendVal, decimals)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return "", err
	}

	hash, err = CallEthTransfer(ctx, chainId, rpc, pri, pub, toAddress, sendValue, gasLimit)
	return
}

func SendEthTokenTransfer(ctx context.Context, chainId uint, pri, pub, toAddress, coin string, sendVal string) (hash string, err error) {
	rpc := constant.GetRPCUrlByNetwork(chainId)
	if rpc == "" {
		err = errors.New("chain not support")
		return
	}

	isSupport, _, contractAddress, decimals := sweepUtils.GetContractInfoByChainIdAndSymbol(chainId, coin)
	if !isSupport {
		return "", errors.New("contract address not found")
	}

	var (
		gasLimit uint64 = 96000
	)

	sendValue, err := utils.FormatToOriginalValue(sendVal, decimals)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("%s -> %s", constant.GetChainName(chainId), err.Error()))
		return "", err
	}

	hash, err = CallTokenTransfer(ctx, chainId, rpc, pri, pub, toAddress, contractAddress, sendValue, gasLimit)
	return
}
