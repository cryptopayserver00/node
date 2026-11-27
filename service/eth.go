package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
	sweepUtils "node/sweep/utils"
	"node/sweep/utils/erc20"
	"node/utils"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gin-gonic/gin"
)

var (
	// txs-chainId-address (v2 invalidates pre-dedupe cache entries)
	constantTransactionHistory = "txs-v2-%d-%s"
	// tx-chainId-hash
	constantOneTransactionHistory = "tx-%d-%s"
	// tx-pending-chainId-hash
	constantOnePendingTransactionHistory = "tx-pending-%d-%s"
)

func memcacheEnabled() bool {
	return global.NODE_MEMCACHE != nil
}

func (n *NService) GetEthPendingTransactions(ctx context.Context, req request.GetEthTransactions) ([]response.ClientTransaction, error) {
	var err error
	var pendings []response.ClientTransaction

	client.URL = constant.GetRPCUrlByChainId(req.ChainId)
	payload := map[string]any{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "txpool_contentFrom",
		"params":  []string{req.Address},
	}
	var rpcGeneralResponse response.RPCGeneralTxpool
	err = client.HTTPPost(ctx, payload, &rpcGeneralResponse)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return nil, err
	}

	for _, txInfo := range rpcGeneralResponse.Result.Pending {
		model, err := n.DecodeEthTransactionForRpc(req.ChainId, req.Address, txInfo)
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			continue
		}

		byteTxs, err := json.Marshal(txInfo)
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			return nil, err
		}

		// save pending tx to nosql
		if memcacheEnabled() {
			err = global.NODE_MEMCACHE.Set(&memcache.Item{
				Key:        fmt.Sprintf(constantOnePendingTransactionHistory, req.ChainId, txInfo.Hash),
				Value:      byteTxs,
				Expiration: 86400,
			})
			if err != nil {
				global.NODE_LOG.Error(err.Error())
				return nil, err
			}
		}

		pendings = append(pendings, model)
	}
	return pendings, nil
}

func (n *NService) GetEthPendingTransaction(req request.GetEthPendingTransaction) (response.RPCTransaction, error) {
	var tx response.RPCTransaction

	if !memcacheEnabled() {
		return tx, nil
	}

	item, err := global.NODE_MEMCACHE.Get(fmt.Sprintf(constantOnePendingTransactionHistory, req.ChainId, req.Hash))
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return tx, nil
	}

	err = json.Unmarshal(item.Value, &tx)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return tx, err
	}

	return tx, nil
}

func (n *NService) GetEthTransactions(c *gin.Context, req request.GetEthTransactions) ([]response.ClientTransaction, error) {
	return n.GetEvmTransactions(c.Request.Context(), req)
}

// GetEvmTransactions is the shared address-history path for ETH + EVM L2s:
// explorers first (Etherscan V2 / Blockscout / Routescan), Alchemy as fallback.
func (n *NService) GetEvmTransactions(ctx context.Context, req request.GetChainTransactions) ([]response.ClientTransaction, error) {
	txs, err := n.GetEvmTransactionsByExplorer(ctx, req.ChainId, req.Address, req.Asset)
	if err == nil {
		return txs, nil
	}
	global.NODE_LOG.Error(fmt.Sprintf("evm explorer txs failed chain=%d: %v", req.ChainId, err))

	alchemyTxs, alchemyErr := n.FetchEvmTransactionsByAlchemy(ctx, req)
	if alchemyErr != nil {
		global.NODE_LOG.Error(fmt.Sprintf("evm alchemy txs failed chain=%d: %v", req.ChainId, alchemyErr))
		return []response.ClientTransaction{}, nil
	}
	return filterAndSortEvmTxs(alchemyTxs, req.Asset), nil
}

func (n *NService) FetchEthTransactionsByAlchemy(ctx context.Context, req request.GetEthTransactions) ([]response.ClientTransaction, error) {
	return n.FetchEvmTransactionsByAlchemy(ctx, req)
}

func (n *NService) FetchEvmTransactionsByAlchemy(ctx context.Context, req request.GetChainTransactions) ([]response.ClientTransaction, error) {
	client.URL = constant.GetAlchemyRPCUrlByChainId(req.ChainId)
	if client.URL == "" {
		return nil, errors.New("alchemy rpc url is empty")
	}

	var fromRpcAlchemyTxs response.RPCAlchemyTransactionDetails
	var toRpcAlchemyTxs response.RPCAlchemyTransactionDetails
	var transfers []response.RPCAlchemyTransactionTransfer
	fromPayload := map[string]any{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "alchemy_getAssetTransfers",
		"params": []map[string]any{
			{
				"fromBlock":        "0x0",
				"toBlock":          "latest",
				"fromAddress":      req.Address,
				"category":         []string{"erc20", "external"},
				"order":            "desc",
				"withMetadata":     true,
				"excludeZeroValue": true,
				"maxCount":         "0x78",
			},
		},
	}

	toPayload := map[string]any{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "alchemy_getAssetTransfers",
		"params": []map[string]any{
			{
				"fromBlock":        "0x0",
				"toBlock":          "latest",
				"toAddress":        req.Address,
				"category":         []string{"erc20", "external"},
				"order":            "desc",
				"withMetadata":     true,
				"excludeZeroValue": true,
				"maxCount":         "0x78",
			},
		},
	}

	if err := client.HTTPPost(ctx, fromPayload, &fromRpcAlchemyTxs); err != nil {
		global.NODE_LOG.Error(err.Error())
		return nil, err
	}

	if err := client.HTTPPost(ctx, toPayload, &toRpcAlchemyTxs); err != nil {
		global.NODE_LOG.Error(err.Error())
		return nil, err
	}

	transfers = append(transfers, fromRpcAlchemyTxs.Result.Transfers...)
	transfers = append(transfers, toRpcAlchemyTxs.Result.Transfers...)

	var saveTxs []response.ClientTransaction
	for _, v := range transfers {
		model, err := n.DecodeEthTransactionForAlchemy(req.ChainId, req.Address, v)
		if err != nil || model.Hash == "" {
			if err != nil {
				global.NODE_LOG.Error(err.Error())
			}
			continue
		}
		saveTxs = append(saveTxs, model)
	}

	// Pending pool decode is ETH-RPC specific; skip on other EVM chains.
	if req.ChainId == constant.ETH_MAINNET || req.ChainId == constant.ETH_SEPOLIA {
		if pendingTxs, err := n.GetEthPendingTransactions(ctx, req); err == nil {
			saveTxs = append(saveTxs, pendingTxs...)
		}
	}

	// from/to Alchemy queries + pending can repeat the same transfer.
	saveTxs = filterAndSortEvmTxs(saveTxs, "")

	if memcacheEnabled() && len(saveTxs) > 0 {
		if byteTxs, err := json.Marshal(saveTxs); err == nil {
			_ = global.NODE_MEMCACHE.Set(&memcache.Item{
				Key:        fmt.Sprintf(constantTransactionHistory, req.ChainId, strings.ToLower(req.Address)),
				Value:      byteTxs,
				Expiration: 300,
			})
		}
	}

	return saveTxs, nil
}

// UpdateEthTransactionsByAlchemy kept for sweep/task call sites; prefer FetchEvmTransactionsByAlchemy.
func (n *NService) UpdateEthTransactionsByAlchemy(ctx context.Context, req request.GetEthTransactions) error {
	_, err := n.FetchEvmTransactionsByAlchemy(ctx, req)
	return err
}

func (n *NService) DecodeEthTransactionForRpc(chainId uint, address string, tx *response.RPCTransaction) (model response.ClientTransaction, err error) {
	model.ChainId = chainId
	model.Address = address
	model.Hash = tx.Hash

	if tx.Input == "0x" {
		model.Asset = constant.ETH
		value, err := utils.HexStringToBigInt(tx.Value)
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			return model, err
		}

		model.Amount = utils.CalculateBalance(value, 18)
		model.Category = "eth"
	} else {
		isSupportContract, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, tx.To)
		if !isSupportContract {
			return model, errors.New("no support the contract")
		}

		model.Asset = contractName

		_, _, _, transactionValue, err := erc20.DecodeERC20TransactionInputData(chainId, tx.Hash, tx.Input)
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			return model, err
		}

		model.Amount = utils.CalculateBalance(transactionValue, decimals)
		model.Category = "erc20"
	}

	_, contractName, _, _ := sweepUtils.GetContractInfo(chainId, tx.To)

	if strings.EqualFold(tx.From, address) && contractName == constant.SWAP {
		model.Type = "Swap"
	} else if strings.EqualFold(tx.From, address) {
		model.Type = "Send"
	} else if strings.EqualFold(tx.To, address) {
		model.Type = "Received"
	} else {
		global.NODE_LOG.Error(tx.Hash)
		return
	}

	model.Status = "Pending"
	model.BlockTimestamp = 0

	return
}

func (n *NService) DecodeEthTransactionForAlchemy(chainId uint, address string, tx response.RPCAlchemyTransactionTransfer) (model response.ClientTransaction, err error) {
	model.ChainId = chainId
	model.Address = address
	model.Hash = tx.Hash

	native := constant.GetNativeAssetByChainId(chainId)
	if isAlchemyNativeTransfer(chainId, tx) {
		model.Asset = native
		model.Category = "external"

		hexString := strings.TrimPrefix(tx.RawContract.Value, "0x")
		bigIntValue, _ := new(big.Int).SetString(hexString, 16)
		model.Amount = utils.CalculateBalance(bigIntValue, 18)
	} else if tx.Category == "erc20" {
		isSupportContract, contractName, contractAddress, decimals := sweepUtils.GetContractInfo(chainId, tx.RawContract.Address)
		if !isSupportContract {
			return
		}
		model.Asset = contractName
		model.ContractAddress = contractAddress
		model.Category = "erc20"

		hexString := strings.TrimPrefix(tx.RawContract.Value, "0x")
		bigIntValue, _ := new(big.Int).SetString(hexString, 16)
		model.Amount = utils.CalculateBalance(bigIntValue, decimals)
	} else {
		global.NODE_LOG.Error(tx.Hash)
		return
	}

	_, contractName, _, _ := sweepUtils.GetContractInfo(chainId, tx.To)

	if strings.EqualFold(tx.From, address) && contractName == constant.SWAP {
		model.Type = "Swap"
	} else if strings.EqualFold(tx.From, address) {
		model.Type = "Send"
	} else if strings.EqualFold(tx.To, address) {
		model.Type = "Received"
	} else {
		global.NODE_LOG.Error(tx.Hash)
		return
	}

	model.Status = "Success"
	if model.Category == "" {
		model.Category = tx.Category
	}
	timestamp, _ := time.Parse(time.RFC3339, tx.Metadata.BlockTimestamp)
	model.BlockTimestamp = int(timestamp.Unix()) * 1000

	return
}

// Alchemy labels native gas differently per chain (ETH/BNB/MATIC/AVAX); normalize to wallet symbols.
func isAlchemyNativeTransfer(chainId uint, tx response.RPCAlchemyTransactionTransfer) bool {
	if strings.EqualFold(tx.Category, "external") {
		return true
	}
	asset := strings.TrimSpace(tx.Asset)
	if asset == "" {
		return false
	}
	native := constant.GetNativeAssetByChainId(chainId)
	if strings.EqualFold(asset, native) {
		return true
	}
	switch chainId {
	case constant.ETH_MAINNET, constant.ETH_SEPOLIA,
		constant.OP_MAINNET, constant.OP_SEPOLIA,
		constant.ARBITRUM_ONE, constant.ARBITRUM_NOVA, constant.ARBITRUM_SEPOLIA,
		constant.BASE_MAINNET, constant.BASE_SEPOLIA:
		return strings.EqualFold(asset, constant.ETH)
	case constant.BSC_MAINNET, constant.BSC_TESTNET:
		return strings.EqualFold(asset, constant.BNB)
	case constant.POL_MAINNET, constant.POL_TESTNET:
		return strings.EqualFold(asset, constant.POL) || strings.EqualFold(asset, constant.MATIC)
	case constant.AVAX_MAINNET, constant.AVAX_TESTNET:
		return strings.EqualFold(asset, constant.AVAX)
	default:
		return false
	}
}
