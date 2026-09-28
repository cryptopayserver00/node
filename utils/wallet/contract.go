package wallet

import (
	"context"
	"math/big"
	"node/sweep/utils/erc20"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func CallWalletTransactionCore(ctx context.Context, chainId uint, rpc, fromPrivateKey, fromPublicKey, toPublicKey string, ethValue *big.Int, data []byte, gasLimit uint64) (hash string, err error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return
	}
	defer client.Close()

	if len(fromPrivateKey) > 2 && fromPrivateKey[:2] == "0x" {
		fromPrivateKey = fromPrivateKey[2:]
	}

	privateKey, err := crypto.HexToECDSA(fromPrivateKey)
	if err != nil {
		return
	}

	fromAddress := common.HexToAddress(fromPublicKey)
	toAddress := common.HexToAddress(toPublicKey)

	// nonce
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return
	}

	// eth_gasPrice
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return
	}

	// eth_maxPriorityFeePerGas
	gasTipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		return
	}

	useChainId, err := client.NetworkID(ctx)
	if err != nil {
		return
	}

	var signedTx *types.Transaction

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:    useChainId,
		Nonce:      nonce,
		GasTipCap:  gasTipCap,
		GasFeeCap:  gasPrice,
		Gas:        gasLimit,
		To:         &toAddress,
		Value:      ethValue,
		Data:       data,
		AccessList: nil,
	})
	signedTx, err = types.SignTx(tx, types.NewLondonSigner(useChainId), privateKey)
	if err != nil {
		return
	}

	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		return
	}

	return signedTx.Hash().Hex(), nil
}

func CallContractCore(ctx context.Context, rpc, contractAddress, contractFunc string, args ...any) (map[string]any, error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	callData, err := erc20.ERC20ABI.Pack(contractFunc, args...)
	if err != nil {
		return nil, err
	}

	ca := common.HexToAddress(contractAddress)

	msg := ethereum.CallMsg{
		To:   &ca,
		Data: callData,
	}

	callResult, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, err
	}

	inputsMap := make(map[string]any)

	err = erc20.ERC20ABI.UnpackIntoMap(inputsMap, contractFunc, callResult)
	if err != nil {
		return nil, err
	}

	return inputsMap, nil
}

func GetEthBalanceByAddress(ctx context.Context, rpc, address string) (balance *big.Int, err error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return
	}
	defer client.Close()

	balance, err = client.BalanceAt(ctx, common.HexToAddress(address), nil)
	if err != nil {
		return
	}

	return balance, nil
}

func GetTransactionReceiptByHash(ctx context.Context, rpc, hash string) (tx *types.Receipt, err error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return
	}
	defer client.Close()

	receipt, err := client.TransactionReceipt(ctx, common.HexToHash(hash))
	if err != nil {
		return
	}

	return receipt, nil
}

func GetTransactionByHash(ctx context.Context, rpc, hash string) (tx *types.Transaction, isPending bool, err error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return
	}
	defer client.Close()

	transaction, isPending, err := client.TransactionByHash(ctx, common.HexToHash((hash)))
	if err != nil {
		return
	}

	return transaction, isPending, nil
}

// nonce
func GetNonce(ctx context.Context, rpc, fromAddress string) (nonce uint64, err error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return
	}
	defer client.Close()

	nonce, err = client.PendingNonceAt(ctx, common.HexToAddress(fromAddress))
	if err != nil {
		return
	}

	return nonce, nil
}

// eth_gasPrice
func GetGasPrice(ctx context.Context, rpc string) (gasPrice *big.Int, err error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return
	}
	defer client.Close()

	gasPrice, err = client.SuggestGasPrice(ctx)
	if err != nil {
		return
	}

	return gasPrice, nil
}

// eth_maxPriorityFeePerGas
func GetGasTipCap(ctx context.Context, rpc string) (gasTipCap *big.Int, err error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return
	}
	defer client.Close()

	gasTipCap, err = client.SuggestGasTipCap(ctx)
	if err != nil {
		return
	}

	return gasTipCap, nil
}

// chainId
func GetChainId(ctx context.Context, rpc string) (chainId *big.Int, err error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return
	}
	defer client.Close()

	chainId, err = client.NetworkID(ctx)
	if err != nil {
		return
	}

	return chainId, nil
}

func EstimateGas(ctx context.Context, rpc, fromAddress, toAddress string, value *big.Int, data []byte) (gas uint64, err error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return
	}
	defer client.Close()

	fromAddressHex := common.HexToAddress(fromAddress)
	toAddressHex := common.HexToAddress(toAddress)

	gas, err = client.EstimateGas(ctx, ethereum.CallMsg{
		From:  fromAddressHex,
		To:    &toAddressHex,
		Value: value,
		Data:  data,
	})

	if err != nil {
		return
	}

	return gas, nil
}
