package response

import "node/model"

// type NetworkInfo struct {
// 	TatumUrl    string `json:"tatum_url"`
// 	TatumKey    string `json:"tatum_key"`
// 	ChainId     uint   `json:"chain_id"`
// 	RPCUrl      string `json:"rpc_url"`
// 	HTTPUrl     string `json:"http_url"`
// 	HTTPKey     string `json:"http_key"`
// }

type NetworkInfo struct {
	ChainId     uint   `json:"chain_id"`
	LatestBlock string `json:"latest_block"`
	CacheBlock  string `json:"cache_block"`
	SweepBlock  string `json:"sweep_block"`
}

type StoreUserWallet struct {
	Address string `json:"address" `
	ChainId uint   `json:"chain_id"`
}

type BulkStoreUserWalletResponse struct {
	BulkStorage []StoreUserWallet `json:"bulk_storage"`
}

type StoreChainContract struct {
	Contract string `json:"contract" `
	ChainId  uint   `json:"chain_id"`
}

type BulkStoreChainContractResponse struct {
	BulkStorage []StoreChainContract `json:"bulk_storage"`
}

type OwnListResponse struct {
	Transactions []model.OwnTransaction `json:"transactions"`
	Total        int64                  `json:"total"`
	Page         int                    `json:"page"`
	PageSize     int                    `json:"pageSize"`
}

type FreeCoinResponse struct {
	ChainId uint   `json:"chain_id"`
	Hash    string `json:"hash"`
}
