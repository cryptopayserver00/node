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
	BlockNumber    int64  `json:"block_number"`
	// TransferIndex 链上明细序号（vout / log_index / instruction_index）。
	// 未设置时由服务端根据 from/to/amount 派生稳定 transfer_key。
	TransferIndex    int  `json:"transfer_index"`
	HasTransferIndex bool `json:"-"`
}

func (r *NotificationRequest) SetTransferIndex(idx int) {
	r.TransferIndex = idx
	r.HasTransferIndex = true
}
