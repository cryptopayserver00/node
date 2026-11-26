package wallet

import (
	"context"
	"fmt"
	"math/big"
	"node/sweep/utils/erc20"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	// Post-Glamsterdam (Sepolia 2026-10-06+): first transfer to a never-used
	// account needs ~204_600 gas. Keep a floor above that when estimate fails.
	evmNativeGasFloor = uint64(250_000)
	evmTokenGasFloor  = uint64(350_000)
	evmGasBufferPct   = uint64(20) // +20% on estimate
	evmReceiptWait    = 60 * time.Second
	evmReceiptPoll    = 2 * time.Second
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

	if ethValue == nil {
		ethValue = big.NewInt(0)
	}

	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return
	}

	useChainId, err := client.NetworkID(ctx)
	if err != nil {
		return
	}

	gasLimit = resolveEvmGasLimit(ctx, client, fromAddress, toAddress, ethValue, data, gasLimit)
	tipCap, feeCap, err := resolveEvmFees(ctx, client)
	if err != nil {
		return
	}

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   useChainId,
		Nonce:     nonce,
		GasTipCap: tipCap,
		GasFeeCap: feeCap,
		Gas:       gasLimit,
		To:        &toAddress,
		Value:     ethValue,
		Data:      data,
	})
	signedTx, err := types.SignTx(tx, types.NewLondonSigner(useChainId), privateKey)
	if err != nil {
		return
	}

	if err = client.SendTransaction(ctx, signedTx); err != nil {
		return
	}

	hash = signedTx.Hash().Hex()

	// Wait briefly so callers (freecoin) don't treat an on-chain failure as success.
	if receipt, waitErr := waitEvmReceipt(ctx, client, signedTx.Hash(), evmReceiptWait); waitErr == nil {
		if receipt.Status != types.ReceiptStatusSuccessful {
			return hash, fmt.Errorf("transaction failed on-chain (status=0, gasUsed=%d/%d): %s", receipt.GasUsed, gasLimit, hash)
		}
	}

	return hash, nil
}

func resolveEvmGasLimit(ctx context.Context, client *ethclient.Client, from, to common.Address, value *big.Int, data []byte, gasLimit uint64) uint64 {
	estimated, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    &to,
		Value: value,
		Data:  data,
	})
	if err == nil && estimated > 0 {
		buffered := estimated + estimated*evmGasBufferPct/100
		if gasLimit == 0 || buffered > gasLimit {
			return buffered
		}
		return gasLimit
	}

	if gasLimit > 0 {
		return gasLimit
	}
	if len(data) == 0 {
		return evmNativeGasFloor
	}
	return evmTokenGasFloor
}

func resolveEvmFees(ctx context.Context, client *ethclient.Client) (tipCap, feeCap *big.Int, err error) {
	tipCap, err = client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, nil, err
	}
	if tipCap == nil || tipCap.Sign() <= 0 {
		tipCap = big.NewInt(1)
	}

	suggest, suggestErr := client.SuggestGasPrice(ctx)
	header, headerErr := client.HeaderByNumber(ctx, nil)

	feeCap = new(big.Int).Set(tipCap)
	if headerErr == nil && header != nil && header.BaseFee != nil {
		// feeCap = 2*baseFee + tip — survives one base-fee bump.
		feeCap = new(big.Int).Mul(header.BaseFee, big.NewInt(2))
		feeCap.Add(feeCap, tipCap)
	}
	if suggestErr == nil && suggest != nil && suggest.Cmp(feeCap) > 0 {
		feeCap = new(big.Int).Set(suggest)
	}
	if feeCap.Cmp(tipCap) < 0 {
		feeCap = new(big.Int).Set(tipCap)
	}
	return tipCap, feeCap, nil
}

func waitEvmReceipt(ctx context.Context, client *ethclient.Client, txHash common.Hash, timeout time.Duration) (*types.Receipt, error) {
	deadline := time.Now().Add(timeout)
	for {
		receipt, err := client.TransactionReceipt(ctx, txHash)
		if err == nil {
			return receipt, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for receipt %s", txHash.Hex())
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(evmReceiptPoll):
		}
	}
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

func GetTokenBalanceByAddress(ctx context.Context, rpc, address, tokenAddress string) (balance *big.Int, err error) {
	return CallTokenBalanceOf(ctx, rpc, address, tokenAddress)
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
