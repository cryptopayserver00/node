package response

type ClientBalanceResponse struct {
	Balance float64 `json:"balance"`
}

type ClientBtcTxResponse struct {
	Hash           string               `json:"hash"`
	Amount         any                  `json:"amount"`
	Status         string               `json:"status"`
	BlockTimestamp int64                `json:"blockTimestamp"`
	Inputs         []ClientBtcTxInputs  `json:"inputs"`
	Outputs        []ClientBtcTxOutputs `json:"outputs"`
	Url            string               `json:"url"`
	Fee            string               `json:"fee"`
	Type           string               `json:"type"`
	Asset          string               `json:"asset"`
}

type ClientBtcTxInputs struct {
	Address string  `json:"address"`
	Amount  float64 `json:"amount"`
}

type ClientBtcTxOutputs struct {
	Address string  `json:"address"`
	Amount  float64 `json:"amount"`
}

type ClientTransaction struct {
	ChainId         uint   `json:"chainId"`
	Address         string `json:"address" `
	Hash            string `json:"hash" `
	Amount          string `json:"amount"`
	Asset           string `json:"asset" `
	ContractAddress string `json:"contractAddress"`
	Type            string `json:"type" `
	Category        string `json:"category"`
	Status          string `json:"status"`
	BlockTimestamp  int    `json:"blockTimestamp" `
}
