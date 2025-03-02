package service

import (
	"errors"
	"node/global"
	"node/global/constant"
	"node/model"
	"node/model/node/request"

	"gorm.io/gorm"
)

// func (n *NService) SaveTx(chainId uint, hash string) (err error) {
// 	if !constant.IsNetworkSupport(chainId) {
// 		return errors.New("do not support network")
// 	}

// 	hasWallet, err := n.HasTxByChainIdAndHash(chainId, hash)
// 	if err != nil {
// 		return
// 	}

// 	if hasWallet {
// 		return nil
// 	}

// 	var saveTx model.Transaction
// 	saveTx.ChainId = chainId
// 	saveTx.Hash = hash
// 	saveTx.Status = 1

// 	if err = global.NODE_DB.Create(&saveTx).Error; err != nil {
// 		return
// 	}

// 	return nil
// }

// func (n *NService) HasTxByChainIdAndHash(chainId uint, hash string) (hasWallet bool, err error) {
// 	var findTx model.Transaction

// 	err = global.NODE_DB.Where("chain_id = ? AND hash = ?", chainId, hash).First(&findTx).Error
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return false, nil
// 		}
// 		return false, err
// 	}

// 	if findTx.ID > 0 {
// 		return true, nil
// 	}

// 	return false, nil
// }

func (n *NService) SaveOwnTx(request request.NotificationRequest) (id uint, err error) {
	if !constant.IsNetworkSupport(request.Chain) {
		return 0, errors.New("do not support network")
	}

	hasWallet, err := n.HasOwnTxByNotificationObj(request)
	if err != nil {
		return
	}

	if hasWallet {
		return 0, nil
	}

	var ownTx model.OwnTransaction
	ownTx.ChainId = request.Chain
	ownTx.Hash = request.Hash
	ownTx.Address = request.Address
	ownTx.FromAddress = request.FromAddress
	ownTx.ToAddress = request.ToAddress
	ownTx.Token = request.Token
	ownTx.TransactType = request.TransactType
	ownTx.Amount = request.Amount
	ownTx.BlockTimestamp = request.BlockTimestamp
	ownTx.Status = 1

	if err = global.NODE_DB.Create(&ownTx).Error; err != nil {
		return 0, err
	}

	return ownTx.ID, nil
}

func (n *NService) HasOwnTxByNotificationObj(request request.NotificationRequest) (hasWallet bool, err error) {
	var findOwnTx model.OwnTransaction

	err = global.NODE_DB.Where("chain_id = ? AND hash = ? AND address = ? AND from_address = ? AND to_address = ? AND token = ? AND transact_type = ? AND amount = ? AND block_timestamp = ?",
		request.Chain, request.Hash, request.Address, request.FromAddress, request.ToAddress, request.Token, request.TransactType, request.Amount, request.BlockTimestamp).First(&findOwnTx).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	if findOwnTx.ID > 0 {
		return true, nil
	}

	return false, nil
}

func (n *NService) GetOwnTxById(id string) (findOwnTx model.OwnTransaction, err error) {
	err = global.NODE_DB.Where("id = ?", id).First(&findOwnTx).Error
	return
}

// func (n *NService) GetTransactionByChainAndHash(chainId uint, hash string) (interface{}, error) {
// 	var findTx model.Transaction

// 	err := global.NODE_DB.Where("chain_id = ? AND hash = ?", chainId, hash).First(&findTx).Error
// 	if err == nil && findTx.ID > 0 {
// 		return findTx, nil
// 	}

// 	return n.getTxByChainAndHash(chainId, hash)
// }

func (n *NService) GetTransactionsByChainAndAddress(req request.TransactionsByChainAndAddress) ([]model.OwnTransaction, int64, error) {
	var txs []model.OwnTransaction

	var total int64
	limit := req.PageSize
	offset := req.PageSize * (req.Page - 1)
	db := global.NODE_DB.Model(&model.OwnTransaction{})

	if req.ChainId != 0 {
		db.Where("chain_id", req.ChainId)
	}

	if req.Address != "" {
		db.Where("from_address = ? OR to_address = ?", req.Address, req.Address)
	}

	if err := db.Count(&total).Order("created_at desc").Offset(offset).Limit(limit).Find(&txs).Error; err != nil {
		return nil, total, err
	}

	return txs, total, nil
}

// func (n *NService) getTxByChainAndHash(chainId uint, hash string) (interface{}, error) {

// 	if constant.IsNetworkLikeEth(chainId) {
// 		return n.handleLikeEthChain(chainId, hash)
// 	}

// 	if constant.IsNetworkLikeTron(chainId) {
// 		return n.handleLinkTronChain(chainId, hash)
// 	}

// 	if constant.IsNetworkLikeBtc(chainId) {
// 		return n.handleLinkBtcChain(chainId, hash)
// 	}

// 	if constant.IsNetworkLikeLtc(chainId) {
// 		return n.handleLinkLtcChain(chainId, hash)
// 	}

// 	return nil, errors.New("network not support")
// }

// func (n *NService) handleLikeEthChain(chainId uint, hash string) (interface{}, error) {
// 	var err error
// 	var tx model.Transaction

// 	client.URL = constant.GetRPCUrlByNetwork(chainId)
// 	var rpcDetail response.RPCTransactionDetail
// 	var jsonRpcRequest request.JsonRpcRequest
// 	jsonRpcRequest.Id = 1
// 	jsonRpcRequest.Jsonrpc = "2.0"
// 	jsonRpcRequest.Method = "eth_getTransactionByHash"
// 	jsonRpcRequest.Params = []interface{}{hash}

// 	err = client.HTTPPost(jsonRpcRequest, &rpcDetail)
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return nil, err
// 	}

// 	var rpcBlockInfo response.RPCBlockInfo
// 	jsonRpcRequest.Id = 1
// 	jsonRpcRequest.Jsonrpc = "2.0"
// 	jsonRpcRequest.Method = "eth_getBlockByNumber"
// 	jsonRpcRequest.Params = []interface{}{rpcDetail.Result.BlockNumber, false}

// 	err = client.HTTPPost(jsonRpcRequest, &rpcBlockInfo)
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return nil, err
// 	}

// 	tx.Hash = rpcDetail.Result.Hash
// 	tx.ChainId = chainId
// 	tx.BlockNumber, _ = utils.HexStringToUint64(rpcDetail.Result.BlockNumber)
// 	tx.BlockHash = rpcDetail.Result.BlockHash
// 	tx.From = rpcDetail.Result.From
// 	tx.To = rpcDetail.Result.To
// 	tx.Gas, _ = utils.HexStringToUint64(rpcDetail.Result.Gas)
// 	tx.GasPrice, _ = utils.HexStringToUint64(rpcDetail.Result.GasPrice)
// 	tx.Input = rpcDetail.Result.Input
// 	tx.MaxFeePerGas, _ = utils.HexStringToUint64(rpcDetail.Result.MaxFeePerGas)
// 	tx.MaxPriorityFeePerGas, _ = utils.HexStringToUint64(rpcDetail.Result.MaxPriorityFeePerGas)
// 	tx.Nonce, _ = utils.HexStringToUint64(rpcDetail.Result.Nonce)
// 	tx.TransactionIndex, _ = utils.HexStringToUint64(rpcDetail.Result.TransactionIndex)
// 	tx.Type, _ = utils.HexStringToUint64(rpcDetail.Result.Type)
// 	bigIntValue, _ := utils.HexStringToBigInt(rpcDetail.Result.Value)
// 	tx.Value = bigIntValue.String()

// 	blockTimeStamp, _ := utils.HexStringToUint64(rpcBlockInfo.Result.Timestamp)
// 	tx.BlockTimestamp = int(blockTimeStamp) * 1000

// 	if err = global.NODE_DB.Create(&tx).Error; err != nil {
// 		return nil, err
// 	}

// 	return tx, nil
// }

// func (n *NService) handleLinkTronChain(chainId uint, hash string) (interface{}, error) {
// 	var err error

// 	client.URL = constant.TronGetTxByIdByNetwork(chainId)
// 	client.Headers = map[string]string{
// 		"TRON-PRO-API-KEY": constant.GetRandomHTTPKeyByNetwork(chainId),
// 	}

// 	var txRequest request.TronGetBlockTxByIdRequest
// 	txRequest.Value = hash
// 	var txResponse response.TronGetTxResponse
// 	err = client.HTTPPost(txRequest, &txResponse)
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return nil, err
// 	}

// 	return txResponse, nil
// }

// func (n *NService) handleLinkBtcChain(chainId uint, hash string) (interface{}, error) {
// 	var err error

// 	client.URL = constant.TatumGetBitcoinTxByHash + hash
// 	client.Headers = map[string]string{
// 		"x-api-key": constant.GetTatumRandomKeyByNetwork(chainId),
// 	}
// 	var bitcoinTxResponse tatum.TatumBitcoinTx
// 	err = client.HTTPGet(&bitcoinTxResponse)
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return nil, err
// 	}

// 	return bitcoinTxResponse, nil
// }

// func (n *NService) handleLinkLtcChain(chainId uint, hash string) (interface{}, error) {
// 	var err error

// 	client.URL = constant.TatumGetLitecoinTxByHash + hash
// 	client.Headers = map[string]string{
// 		"x-api-key": constant.GetTatumRandomKeyByNetwork(chainId),
// 	}

// 	var litecoinTxResponse tatum.TatumLitecoinTx
// 	err = client.HTTPGet(&litecoinTxResponse)
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return nil, err
// 	}

// 	return litecoinTxResponse, nil
// }
