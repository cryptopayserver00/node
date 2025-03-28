package response

type XrpscanBlockResponse struct {
	Result XrpscanResult `json:"result"`
}

type XrpscanResult struct {
	Ledger      XrpscanLedger `json:"ledger"`
	LedgerHash  string        `json:"ledger_hash"`
	LedgerIndex int           `json:"ledger_index"`
	Status      string        `json:"status"`
	Validated   bool          `json:"validated"`
}

type XrpscanLedger struct {
	AccountHash         string               `json:"account_hash"`
	CloseFlags          int                  `json:"close_flags"`
	CloseTime           int                  `json:"close_time"`
	CloseTimeHuman      string               `json:"close_time_human"`
	CloseTimeIso        string               `json:"close_time_iso"`
	CloseTimeResolution int                  `json:"close_time_resolution"`
	Closed              bool                 `json:"closed"`
	LedgerHash          string               `json:"ledger_hash"`
	LedgerIndex         int                  `json:"ledger_index"`
	ParentCloseTime     int                  `json:"parent_close_time"`
	ParentHash          string               `json:"parent_hash"`
	TotalCoins          string               `json:"total_coins"`
	TransactionHash     string               `json:"transaction_hash"`
	Transactions        []XrpscanTransaction `json:"transactions"`
}

type XrpscanTransactionResponse struct {
	Result XrpscanTransaction `json:"result"`
}

type XrpscanTransaction struct {
	CloseTimeIso string                   `json:"close_time_iso"`
	Hash         string                   `json:"hash"`
	LedgerHash   string                   `json:"ledger_hash"`
	LedgerIndex  int                      `json:"ledger_index"`
	Meta         XrpscanTransactionMeta   `json:"meta"`
	Status       string                   `json:"status"`
	TxJson       XrpscanTransactionTxJson `json:"tx_json"`
	Validated    bool                     `json:"validated"`
}

type XrpscanTransactionMeta struct {
	// AffectedNodes
	TransactionIndex  int         `json:"TransactionIndex"`
	TransactionResult string      `json:"TransactionResult"`
	DeliveredAmount   interface{} `json:"delivered_amount"`
}

type XrpscanTransactionTxJson struct {
	Account            string      `json:"Account"`
	DeliverMax         interface{} `json:"DeliverMax"`
	Destination        string      `json:"Destination"`
	DestinationTag     int         `json:"DestinationTag"`
	Fee                string      `json:"Fee"`
	Flags              int         `json:"Flags"`
	LastLedgerSequence int         `json:"LastLedgerSequence"`
	Sequence           int         `json:"Sequence"`
	SigningPubKey      string      `json:"SigningPubKey"`
	TicketSequence     int         `json:"TicketSequence"`
	TransactionType    string      `json:"TransactionType"`
	TxnSignature       string      `json:"TxnSignature"`
	Date               int         `json:"date"`
	LedgerIndex        int         `json:"ledgerIndex"`
}
