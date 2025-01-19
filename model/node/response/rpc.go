package response

type RPCBlockInfo struct {
	JsonRpc string             `json:"jsonrpc"`
	Id      int                `json:"id"`
	Result  RPCBlockInfoResult `json:"result"`
}

type RPCBlockInfoResult struct {
	BaseFeePerGas    string   `json:"baseFeePerGas"`
	Difficulty       string   `json:"difficulty"`
	ExtraData        string   `json:"extraData"`
	GasLimit         string   `json:"gasLimit"`
	GasUsed          string   `json:"gasUsed"`
	Hash             string   `json:"hash"`
	LogsBloom        string   `json:"logsBloom"`
	Miner            string   `json:"miner"`
	MixHash          string   `json:"mixHash"`
	Nonce            string   `json:"nonce"`
	Number           string   `json:"number"`
	ParentHash       string   `json:"parentHash"`
	ReceiptsRoot     string   `json:"receiptsRoot"`
	Sha3Uncles       string   `json:"sha3Uncles"`
	Size             string   `json:"size"`
	StateRoot        string   `json:"stateRoot"`
	Timestamp        string   `json:"timestamp"`
	TotalDifficulty  string   `json:"totalDifficulty"`
	Transactions     []string `json:"transactions"`
	TransactionsRoot string   `json:"transactionsRoot"`
	Uncles           []string `json:"uncles"`
}

type RPCBlockDetail struct {
	JsonRpc string               `json:"jsonrpc"`
	Id      int                  `json:"id"`
	Result  RPCBlockDetailResult `json:"result"`
}

type RPCBlockDetailResult struct {
	BaseFeePerGas    string           `json:"baseFeePerGas"`
	Difficulty       string           `json:"difficulty"`
	ExtraData        string           `json:"extraData"`
	GasLimit         string           `json:"gasLimit"`
	GasUsed          string           `json:"gasUsed"`
	Hash             string           `json:"hash"`
	LogsBloom        string           `json:"logsBloom"`
	Miner            string           `json:"miner"`
	MixHash          string           `json:"mixHash"`
	Nonce            string           `json:"nonce"`
	Number           string           `json:"number"`
	ParentHash       string           `json:"parentHash"`
	ReceiptsRoot     string           `json:"receiptsRoot"`
	Sha3Uncles       string           `json:"sha3Uncles"`
	Size             string           `json:"size"`
	StateRoot        string           `json:"stateRoot"`
	Timestamp        string           `json:"timestamp"`
	TotalDifficulty  string           `json:"totalDifficulty"`
	Transactions     []RPCTransaction `json:"transactions"`
	TransactionsRoot string           `json:"transactionsRoot"`
	Uncles           []string         `json:"uncles"`
}

type RPCTransactionDetail struct {
	JsonRpc string         `json:"jsonrpc"`
	Id      int            `json:"id"`
	Result  RPCTransaction `json:"result"`
}

type RPCTransaction struct {
	BlockHash            string `json:"blockHash"`
	BlockNumber          string `json:"blockNumber"`
	Hash                 string `json:"hash"`
	From                 string `json:"from"`
	Gas                  string `json:"gas"`
	GasPrice             string `json:"gasPrice"`
	Input                string `json:"input"`
	MaxFeePerGas         string `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas"`
	Nonce                string `json:"nonce"`
	R                    string `json:"r"`
	S                    string `json:"s"`
	SourceHash           string `json:"sourceHash"`
	To                   string `json:"to"`
	TransactionIndex     string `json:"transactionIndex"`
	Type                 string `json:"type"`
	V                    string `json:"v"`
	Value                string `json:"value"`
	ChainId              string `json:"chainId"`
}

type RPCReceiptTransactionDetail struct {
	JsonRpc string                `json:"jsonrpc"`
	Id      int                   `json:"id"`
	Result  RPCReceiptTransaction `json:"result"`
}

type RPCReceiptTransaction struct {
	TransactionHash   string           `json:"transactionHash"`
	BlockHash         string           `json:"blockHash"`
	BlockNumber       string           `json:"blockNumber"`
	LogsBloom         string           `json:"logsBloom"`
	GasUsed           string           `json:"gasUsed"`
	ContractAddress   string           `json:"contractAddress"`
	CumulativeGasUsed string           `json:"cumulativeGasUsed"`
	TransactionIndex  string           `json:"transactionIndex"`
	From              string           `json:"from"`
	To                string           `json:"to"`
	Type              string           `json:"type"`
	EffectiveGasPrice string           `json:"effectiveGasPrice"`
	Logs              []RPCReceiptLogs `json:"logs"`
	Status            string           `json:"status"`
}

type RPCReceiptLogs struct {
	BlockHash        string   `json:"blockHash"`
	Address          string   `json:"address"`
	LogIndex         string   `json:"logIndex"`
	Data             string   `json:"data"`
	Removed          bool     `json:"removed"`
	Topics           []string `json:"topics"`
	BlockNumber      string   `json:"blockNumber"`
	TransactionIndex string   `json:"transactionIndex"`
	TransactionHash  string   `json:"transactionHash"`
}

type RPCGeneralTxpool struct {
	JsonRpc string       `json:"jsonrpc"`
	Id      int          `json:"id"`
	Result  TxPoolResult `json:"result"`
}

type TxPoolResult struct {
	Pending map[string]*RPCTransaction `json:"pending"`
	Queued  map[string]interface{}     `json:"queued"`
}

type RPCGeneral struct {
	JsonRpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  string `json:"result"`
}

type RPCStringArrayGeneral struct {
	JsonRpc string   `json:"jsonrpc"`
	Id      int      `json:"id"`
	Result  []string `json:"result"`
}

type RPCAlchemyTransactionDetails struct {
	JsonRpc string                `json:"jsonrpc"`
	Id      int                   `json:"id"`
	Result  RPCAlchemyTransaction `json:"result"`
}

type RPCAlchemyTransaction struct {
	PageKey   string                          `json:"pageKey"`
	Transfers []RPCAlchemyTransactionTransfer `json:"transfers"`
}

type RPCAlchemyTransactionTransfer struct {
	Category    string                                   `json:"category"`
	BlockNum    string                                   `json:"blockNum"`
	From        string                                   `json:"from"`
	To          string                                   `json:"to"`
	Value       float64                                  `json:"value"`
	TokenId     string                                   `json:"tokenId"`
	Asset       string                                   `json:"asset"`
	UniqueId    string                                   `json:"uniqueId"`
	Hash        string                                   `json:"hash"`
	RawContract RPCAlchemyTransactionTransferRawContract `json:"rawContract"`
	Metadata    RPCAlchemyTransactionTransferMetadata    `json:"metadata"`
}

type RPCAlchemyTransactionTransferRawContract struct {
	Value   string `json:"value"`
	Address string `json:"address"`
	Decimal string `json:"decimal"`
}

type RPCAlchemyTransactionTransferMetadata struct {
	BlockTimestamp string `json:"blockTimestamp"`
}

type CallResult struct {
	From         string       `json:"from"`
	Gas          string       `json:"gas"`
	GasUsed      string       `json:"gasUsed"`
	Input        string       `json:"input"`
	Output       string       `json:"output"`
	To           string       `json:"to"`
	Type         string       `json:"type"`
	Value        string       `json:"value"`
	Error        string       `json:"error"`
	RevertReason string       `json:"revertReason"`
	Calls        []CallResult `json:"calls"`
}

type RPCInnerTxInfo struct {
	JsonRpc string     `json:"jsonrpc"`
	Id      int        `json:"id"`
	Result  CallResult `json:"result"`
}

type RPCBlockInnerDetail struct {
	JsonRpc string                   `json:"jsonrpc"`
	Id      int                      `json:"id"`
	Result  []BlockInnerDetailResult `json:"result"`
}

type BlockInnerDetailResult struct {
	Hash   string     `json:"txHash"`
	Result CallResult `json:"result"`
}
