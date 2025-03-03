package request

type StoreUserWallet struct {
	Address string `json:"address" form:"address" binding:"required"`
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
}

type BulkStoreUserWallet struct {
	BulkStorage []StoreUserWallet `json:"bulk_storage" form:"bulk_storage" binding:"required"`
}

type GetNetworkInfo struct {
	ChainId uint `json:"chain_id" form:"chain_id" binding:"required"`
}

type StoreChainContract struct {
	ChainId  uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Symbol   string `json:"symbol" form:"symbol" binding:"required"`
	Decimals int    `json:"decimals" form:"decimals" binding:"required"`
	Contract string `json:"contract" form:"contract" binding:"required"`
}

type BulkStoreChainContract struct {
	BulkStorage []StoreChainContract `json:"bulk_storage" form:"bulk_storage" binding:"required"`
}

type TransactionByChainAndHash struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Hash    string `json:"hash" form:"hash" binding:"required"`
}

type TransactionsByChainAndAddress struct {
	ChainIds  string `json:"chainids" form:"chainids"`
	Addresses string `json:"addresses" form:"addresses"`
	Page      int    `json:"page" form:"page"`
	PageSize  int    `json:"pageSize" form:"page_size"`
}

type SendMessageToTelegram struct {
	AuthKey string `json:"auth_key" form:"auth_key" binding:"required"`
	Message string `json:"message" form:"message" binding:"required"`
}

type RevokeTelegramKey struct {
	AuthKey string `json:"auth_key" form:"auth_key" binding:"required"`
}

// eth
type GetEthTransactions struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
	Asset   string `json:"asset" form:"asset"`
}

// eth
type GetEthPendingTransaction struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Hash    string `json:"hash" form:"hash" binding:"required"`
}

// bsc
type GetBscTransactions struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
	Asset   string `json:"asset" form:"asset"`
}

// btc
type GetBtcBalance struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
}

type GetBtcFeeRate struct {
	ChainId uint `json:"chain_id" form:"chain_id" binding:"required"`
}

type GetBtcAddressUtxo struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
}

type PostBtcBroadcast struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	TxData  string `json:"tx_data" form:"tx_data" binding:"required"`
}

type GetBtcTransactions struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
}

type GetBtcTransactionDetail struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Hash    string `json:"hash" form:"hash" binding:"required"`
}

// ltc
type GetLtcBalance struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
}

type GetLtcFeeRate struct {
	ChainId uint `json:"chain_id" form:"chain_id" binding:"required"`
}

type PostLtcBroadcast struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	TxData  string `json:"tx_data" form:"tx_data" binding:"required"`
}

type GetLtcTransactions struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
}

type GetLtcTxByHash struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Hash    string `json:"hash" form:"hash" binding:"required"`
}

type GetLtcAddressUtxo struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
}

type GetTronTransactions struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
}

type GetTrxTransactions struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
}

type GetTrc20Transactions struct {
	ChainId         uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address         string `json:"address" form:"address" binding:"required"`
	ContractAddress string `json:"contract_address" form:"contract_address" binding:"required"`
}

type GetSolanaTransactions struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
}

type GetSolTransactions struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
}

type GetSplTransactions struct {
	ChainId         uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address         string `json:"address" form:"address" binding:"required"`
	ContractAddress string `json:"contract_address" form:"contract_address" binding:"required"`
}

type GetTonTransactions struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
}

type GetTonCoinTransactions struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
}

type GetTon20Transactions struct {
	ChainId         uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address         string `json:"address" form:"address" binding:"required"`
	ContractAddress string `json:"contract_address" form:"contract_address" binding:"required"`
}

type GetFreeCoin struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
	Coin    string `json:"coin" form:"coin" binding:"required"`
	Amount  string `json:"amount" form:"amount"`
}

// xrp
type GetXrpTransactions struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
	Asset   string `json:"asset" form:"asset"`
}

// bch
type GetBchTransactions struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
	Asset   string `json:"asset" form:"asset"`
}

// arb
type GetArbTransactions struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
	Asset   string `json:"asset" form:"asset"`
}

// avax
type GetAvaxTransactions struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
	Asset   string `json:"asset" form:"asset"`
}

// pol
type GetPolTransactions struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
	Asset   string `json:"asset" form:"asset"`
}

// base
type GetBaseTransactions struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
	Asset   string `json:"asset" form:"asset"`
}

// op
type GetOpTransactions struct {
	ChainId uint   `json:"chain_id" form:"chain_id" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
	Asset   string `json:"asset" form:"asset"`
}
