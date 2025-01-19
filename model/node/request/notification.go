package request

type NotificationRequest struct {
	Hash           string `json:"hash"`
	Address        string `json:"address"`
	FromAddress    string `json:"from_address"`
	ToAddress      string `json:"to_address"`
	Token          string `json:"token"`
	TransactType   string `json:"transact_type"`
	Chain          uint   `json:"chain"`
	Amount         string `json:"amount"`
	BlockTimestamp int    `json:"block_timestamp"`
}
