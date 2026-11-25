package request

type GetCryptoPrice struct {
	// ids: comma-separated CoinGecko ids, e.g. "bitcoin,ethereum"
	Ids string `json:"ids" form:"ids"`
	// currency: vs currency, default "usd"
	Currency string `json:"currency" form:"currency"`
	// format: "gecko" (default, CoinGecko id keys) | "symbol" (btc/eth/... for deshop)
	Format string `json:"format" form:"format"`
}
