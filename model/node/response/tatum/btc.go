package tatum

type TatumGetBitcoinInfo struct {
	Chain         string  `json:"chain"`
	Blocks        int     `json:"blocks"`
	Headers       int     `json:"headers"`
	Bestblockhash string  `json:"bestblockhash"`
	Difficulty    float64 `json:"difficulty"`
}

type TatumGetBitcoinBlock struct {
	Hash       string           `json:"hash"`
	Height     int              `json:"height"`
	Depth      int              `json:"depth"`
	Version    int              `json:"version"`
	PrevBlock  string           `json:"prevBlock"`
	MerkleRoot string           `json:"merkleRoot"`
	Time       int              `json:"time"`
	Bits       int              `json:"bits"`
	Nonce      int              `json:"nonce"`
	Txs        []TatumBitcoinTx `json:"txs"`
}

type TatumBitcoinTx struct {
	Hash        string            `json:"hash"`
	Hex         string            `json:"hex"`
	WitnessHash string            `json:"witnessHash"`
	Fee         int               `json:"fee"`
	Rate        int               `json:"rate"`
	Mtime       int               `json:"mtime"`
	BlockNumber int               `json:"blockNumber"`
	Block       string            `json:"block"`
	Time        int               `json:"time"`
	Index       int               `json:"index"`
	Version     int               `json:"version"`
	Locktime    int               `json:"locktime"`
	Inputs      []BitcoinTxInput  `json:"inputs"`
	Outputs     []BitcoinTxOutput `json:"outputs"`
}

type BitcoinTxInput struct {
	Prevout  BitcoinTxInputPrevout `json:"prevout"`
	Script   string                `json:"script"`
	Witness  string                `json:"witness"`
	Sequence int                   `json:"sequence"`
	Coin     BitcoinTxInputCoin    `json:"coin"`
}

type BitcoinTxInputPrevout struct {
	Hash  string `json:"hash"`
	Index int    `json:"index"`
}

type BitcoinTxInputCoin struct {
	Version     int    `json:"version"`
	BlockNumber int    `json:"blockNumber"`
	Value       int    `json:"value"`
	Script      string `json:"script"`
	Address     string `json:"address"`
	Coinbase    bool   `json:"coinbase"`
}

type BitcoinTxOutput struct {
	Value   int    `json:"value"`
	Script  string `json:"script"`
	Address string `json:"address"`
}

type BitcoinBalance struct {
	Incoming        string `json:"incoming"`
	IncomingPending string `json:"incomingPending"`
	Outgoing        string `json:"outgoing"`
	OutgoingPending string `json:"outgoingPending"`
}

type BitcoinFeeRate struct {
	Fast   float64 `json:"fast"`
	Medium float64 `json:"medium"`
	Slow   float64 `json:"slow"`
}

type BitcoinBroadcast struct {
	TxId string `json:"txId"`
}
