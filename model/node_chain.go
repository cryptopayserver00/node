package model

var ChainList []ChainInfo

type Coin struct {
	Symbol     string `json:"symbol"`
	Decimals   int    `json:"decimals"`
	Contract   string `json:"contract"`
	IsMainCoin bool   `json:"isMainCoin"`
}

type ChainInfo struct {
	Name      string `json:"name"`
	Chain     string `json:"chain"`
	ChainId   uint   `json:"chainId"`
	NetworkId int    `json:"networkId"`
	Coins     []Coin `json:"coins"`
}
