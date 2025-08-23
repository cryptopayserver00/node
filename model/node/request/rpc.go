package request

type JsonRpcRequest struct {
	Id      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

// type JsonRpcAlchemyRequest struct {
// 	Id      int                           `json:"id"`
// 	Jsonrpc string                        `json:"jsonrpc"`
// 	Method  string                        `json:"method"`
// 	Params  []JsonRpcAlchemyRequestParams `json:"params"`
// }

// type JsonRpcAlchemyRequestParams struct {
// 	FromBlock        string   `json:"fromBlock"`
// 	ToBlock          string   `json:"toBlock"`
// 	FromAddress      string   `json:"-"`
// 	ToAddress        string   `json:"toAddress"`
// 	Category         []string `json:"category"`
// 	Order            string   `json:"order"`
// 	WithMetadata     bool     `json:"withMetadata"`
// 	ExcludeZeroValue bool     `json:"excludeZeroValue"`
// 	MaxCount         string   `json:"maxCount"`
// }

type XrpJsonRpcRequest struct {
	Method string           `json:"method"`
	Params []map[string]any `json:"params"`
}
