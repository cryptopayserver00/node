package response

type TronGetBlockResponse struct {
	BlockId     string             `json:"blockID"`
	BlockHeader TronGetBlockHedaer `json:"block_header"`
}

type TronGetBlockHedaer struct {
	RawData          TronGetBlockHeaderRawData `json:"raw_data"`
	WitnessSignature string                    `json:"witness_signature"`
}

type TronGetBlockHeaderRawData struct {
	Number          int    `json:"number"`
	TxTrieRoot      string `json:"txTrieRoot"`
	Witness_address string `json:"witness_address"`
	ParentHash      string `json:"parentHash"`
	Version         int    `json:"version"`
	Timestamp       int    `json:"timestamp"`
}

type TronGetBlockByNumResponse struct {
	BlockId      string              `json:"blockID"`
	BlockHeader  TronGetBlockHedaer  `json:"block_header"`
	Transactions []TronGetTxResponse `json:"transactions"`
}

type TronGetTxResponse struct {
	TxStatus       string      `json:"type_status"`
	Ret            []TronTxRet `json:"ret"`
	Signature      []string    `json:"signature"`
	TxID           string      `json:"txID"`
	RawData        TxRawData   `json:"raw_data"`
	RawDataHex     string      `json:"raw_data_hex"`
	BlockNumber    int         `json:"blockNumber"`
	BlockTimestamp int         `json:"block_timestamp"`
}

type TronTxRet struct {
	ContractRet string `json:"contractRet"`
}

type TxRawData struct {
	Contract      []TxRawDataContract `json:"contract"`
	RefBlockBytes string              `json:"ref_block_bytes"`
	RefBlockHash  string              `json:"ref_block_hash"`
	Expiration    int                 `json:"expiration"`
	FeeLimit      int                 `json:"fee_limit"`
	Timestamp     int                 `json:"timestamp"`
}

type TxRawDataContract struct {
	Parameter TxRawDataContractParameter `json:"parameter"`
	Type      string                     `json:"type"`
}

type TxRawDataContractParameter struct {
	Value   TxRawDataContractParameterValue `json:"value"`
	TypeUrl string                          `json:"type_url"`
}

type TxRawDataContractParameterValue struct {
	Amount          int    `json:"amount"`
	OwnerAddress    string `json:"owner_address"`
	ToAddress       string `json:"to_address"`
	Data            string `json:"data"`
	ContractAddress string `json:"contract_address"`
}

type TronValidateAddressResponse struct {
	Result  bool   `json:"result"`
	Message string `json:"message"`
}

type TronContractResponse struct {
	Bytecode        string `json:"bytecode"`
	Name            string `json:"message"`
	ContractAddress string `json:"contract_address"`
}

type TronTxResponse struct {
	Data    []TronGetTxResponse `json:"data"`
	Success bool                `json:"success"`
	Meta    TronTxResponseMete  `json:"meta"`
}

type TronTxResponseMete struct {
	At       int `json:"at"`
	PageSize int `json:"page_size"`
}

type TronTrc20Response struct {
	Data    []TronGetTrc20Response `json:"data"`
	Success bool                   `json:"success"`
	Meta    TronTxResponseMete     `json:"meta"`
}

type TronGetTrc20Response struct {
	TxStatus       string                `json:"type_status"`
	TransactionId  string                `json:"transaction_id"`
	TokenInfo      TronGetTrc20TokenInfo `json:"token_info"`
	BlockTimestamp int                   `json:"block_timestamp"`
	From           string                `json:"from"`
	To             string                `json:"to"`
	Type           string                `json:"type"`
	Value          string                `json:"value"`
}

type TronGetTrc20TokenInfo struct {
	Symbol   string `json:"symbol"`
	Address  string `json:"address"`
	Decimals int    `json:"decimals"`
	Name     string `json:"name"`
}
