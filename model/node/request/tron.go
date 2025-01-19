package request

type TronGetBlockRequest struct {
	IdOrNum string `json:"id_or_num"`
	Detail  bool   `json:"detail"`
}

type TronGetBlockByNumRequest struct {
	Num int `json:"num"`
}

type TronGetBlockTxByIdRequest struct {
	Value string `json:"value"`
}

type TronValidateAddressRequest struct {
	Address string `json:"address"`
	Visible bool   `json:"visible"`
}

type TronContractRequest struct {
	Value   string `json:"value"`
	Visible bool   `json:"visible"`
}
