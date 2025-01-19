package blockstream

type UtxoResponse struct {
	Txid   string     `json:"txid"`
	Vout   int        `json:"vout"`
	Status UtxoStatus `json:"status"`
	Value  int64      `json:"value"`
}

type UtxoStatus struct {
	Confirmed   bool   `json:"confirmed"`
	BlockHeight int    `json:"block_height"`
	BlockHash   string `json:"block_hash"`
	BlockTime   int    `json:"block_time"`
}

type BroadcastResponse struct {
	Txid string `json:"txid"`
}

type TransactionResponse struct {
	Txid     string               `json:"txid"`
	Version  int                  `json:"version"`
	Locktime int                  `json:"locktime"`
	Vin      []TransactionVin     `json:"vin"`
	Vout     []TransactionPrevout `json:"vout"`
	Size     int                  `json:"size"`
	Weight   int                  `json:"weight"`
	Fee      int                  `json:"fee"`
	Status   TransactionStatus    `json:"status"`
}

type TransactionVin struct {
	Txid         string             `json:"txid"`
	Vout         int                `json:"vout"`
	Prevout      TransactionPrevout `json:"prevout"`
	Scriptsig    string             `json:"scriptsig"`
	ScriptsigAsm string             `json:"scriptsig_asm"`
	Witness      []string           `json:"witness"`
	IsCoinbase   bool               `json:"is_coinbase"`
	Sequence     int                `json:"sequence"`
}

type TransactionPrevout struct {
	Scriptpubkey        string `json:"scriptpubkey"`
	ScriptpubkeyAsm     string `json:"scriptpubkey_asm"`
	ScriptpubkeyType    string `json:"scriptpubkey_type"`
	ScriptpubkeyAddress string `json:"scriptpubkey_address"`
	Value               int    `json:"value"`
}

type TransactionStatus struct {
	Confirmed   bool   `json:"confirmed"`
	BlockHeight int    `json:"block_height"`
	BlockHash   string `json:"block_hash"`
	BlockTime   int    `json:"block_time"`
}
