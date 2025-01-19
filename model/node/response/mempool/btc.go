package mempool

type MempoolTx struct {
	TxId     string          `json:"txid"`
	Version  int             `json:"version"`
	LockTime int             `json:"locktime"`
	Vin      []MempoolTxVin  `json:"vin"`
	Vout     []MempoolTxVout `json:"vout"`
	Size     int             `json:"size"`
	Weight   int             `json:"weight"`
	Sigops   int             `json:"sigops"`
	Fee      int             `json:"fee"`
	Status   MempoolTxStatus `json:"status"`
}

type MempoolTxVin struct {
	TxId          string        `json:"txid"`
	Vout          int           `json:"vout"`
	Prevout       MempoolTxVout `json:"prevout"`
	Scriptsig     string        `json:"scriptsig"`
	Scriptsig_asm string        `json:"scriptsig_asm"`
	Witness       []string      `json:"witness"`
	IsCoinbase    bool          `json:"is_coinbase"`
	Sequence      int           `json:"sequence"`
}

type MempoolTxVout struct {
	Scriptpubkey         string `json:"scriptpubkey"`
	Scriptpubkey_asm     string `json:"scriptpubkey_asm"`
	Scriptpubkey_type    string `json:"scriptpubkey_type"`
	Scriptpubkey_address string `json:"scriptpubkey_address"`
	Value                int    `json:"value"`
}

type MempoolTxStatus struct {
	Confirmed  bool   `json:"confirmed"`
	Blockeight int    `json:"block_height"`
	BlockHash  string `json:"block_hash"`
	BlockTime  int    `json:"block_time"`
}

type MempoolBlock struct {
	Id                string  `json:"id"`
	Height            int     `json:"height"`
	Version           int     `json:"version"`
	Timestamp         int     `json:"timestamp"`
	TxCount           int     `json:"tx_count"`
	Size              int     `json:"size"`
	Weight            int     `json:"weight"`
	MerkleRoot        string  `json:"merkle_root"`
	Previousblockhash string  `json:"previousblockhash"`
	Mediantime        int     `json:"mediantime"`
	Nonce             int     `json:"nonce"`
	Bits              int     `json:"bits"`
	Difficulty        float64 `json:"difficulty"`
}

type MempoolUtxo struct {
	TxId   string          `json:"txid"`
	Vout   int             `json:"vout"`
	Status MempoolTxStatus `json:"status"`
	Value  int64           `json:"value"`
}

type MempoolFees struct {
	FastestFee  int `json:"fastestFee"`
	HalfHourFee int `json:"halfHourFee"`
	HourFee     int `json:"hourFee"`
	EconomyFee  int `json:"economyFee"`
	MinimumFee  int `json:"minimumFee"`
}
