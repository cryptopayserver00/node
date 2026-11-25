package response

type CoingeckoPriceCore struct {
	USD           float64 `json:"usd"`
	USDMarketCap  float64 `json:"usd_market_cap"`
	USD24hVol     float64 `json:"usd_24h_vol"`
	USD24hChange  float64 `json:"usd_24h_change"`
	LastUpdatedAt int     `json:"last_updated_at"`
}

// CoingeckoPrice is the CoinGecko simple/price shape (id keys).
type CoingeckoPrice struct {
	BTC   CoingeckoPriceCore `json:"bitcoin"`
	ETH   CoingeckoPriceCore `json:"ethereum"`
	USDT  CoingeckoPriceCore `json:"tether"`
	BNB   CoingeckoPriceCore `json:"binancecoin"`
	SOL   CoingeckoPriceCore `json:"solana"`
	USDC  CoingeckoPriceCore `json:"usd-coin"`
	XRP   CoingeckoPriceCore `json:"ripple"`
	TON   CoingeckoPriceCore `json:"the-open-network"`
	DOGE  CoingeckoPriceCore `json:"dogecoin"`
	ADA   CoingeckoPriceCore `json:"cardano"`
	TRX   CoingeckoPriceCore `json:"tron"`
	AVAX  CoingeckoPriceCore `json:"avalanche-2"`
	SHIB  CoingeckoPriceCore `json:"shiba-inu"`
	DOT   CoingeckoPriceCore `json:"polkadot"`
	LINK  CoingeckoPriceCore `json:"chainlink"`
	BCH   CoingeckoPriceCore `json:"bitcoin-cash"`
	DAI   CoingeckoPriceCore `json:"dai"`
	LTC   CoingeckoPriceCore `json:"litecoin"`
	POL   CoingeckoPriceCore `json:"polygon-ecosystem-token"`
	UNI   CoingeckoPriceCore `json:"uniswap"`
	PEPE  CoingeckoPriceCore `json:"pepe"`
	FIL   CoingeckoPriceCore `json:"filecoin"`
	ARB   CoingeckoPriceCore `json:"arbitrum"`
	OP    CoingeckoPriceCore `json:"optimism"`
	WIF   CoingeckoPriceCore `json:"dogwifcoin"`
	NOT   CoingeckoPriceCore `json:"notcoin"`
	BONK  CoingeckoPriceCore `json:"bonk"`
	AAVE  CoingeckoPriceCore `json:"aave"`
	FLOKI CoingeckoPriceCore `json:"floki"`
	JUP   CoingeckoPriceCore `json:"jupiter"`
	ENS   CoingeckoPriceCore `json:"ethereum-name-service"`
	W     CoingeckoPriceCore `json:"wormhole"`
	MEW   CoingeckoPriceCore `json:"meow"`
	WBTC  CoingeckoPriceCore `json:"wrapped-bitcoin"`
	WETH  CoingeckoPriceCore `json:"weth"`
	WSOL  CoingeckoPriceCore `json:"wrapped-solana"`
	SAND  CoingeckoPriceCore `json:"the-sandbox"`
}

// CustomCoingeckoPrice is the symbol-keyed shape used by deshopserver.
type CustomCoingeckoPrice struct {
	BTC   CoingeckoPriceCore `json:"btc"`
	ETH   CoingeckoPriceCore `json:"eth"`
	USDT  CoingeckoPriceCore `json:"usdt"`
	BNB   CoingeckoPriceCore `json:"bnb"`
	SOL   CoingeckoPriceCore `json:"sol"`
	USDC  CoingeckoPriceCore `json:"usdc"`
	XRP   CoingeckoPriceCore `json:"xrp"`
	TON   CoingeckoPriceCore `json:"ton"`
	DOGE  CoingeckoPriceCore `json:"doge"`
	ADA   CoingeckoPriceCore `json:"ada"`
	TRX   CoingeckoPriceCore `json:"trx"`
	AVAX  CoingeckoPriceCore `json:"avax"`
	SHIB  CoingeckoPriceCore `json:"shib"`
	DOT   CoingeckoPriceCore `json:"dot"`
	LINK  CoingeckoPriceCore `json:"link"`
	BCH   CoingeckoPriceCore `json:"bch"`
	DAI   CoingeckoPriceCore `json:"dai"`
	LTC   CoingeckoPriceCore `json:"ltc"`
	POL   CoingeckoPriceCore `json:"pol"`
	UNI   CoingeckoPriceCore `json:"uni"`
	PEPE  CoingeckoPriceCore `json:"pepe"`
	FIL   CoingeckoPriceCore `json:"fil"`
	ARB   CoingeckoPriceCore `json:"arb"`
	OP    CoingeckoPriceCore `json:"op"`
	WIF   CoingeckoPriceCore `json:"wif"`
	NOT   CoingeckoPriceCore `json:"not"`
	BONK  CoingeckoPriceCore `json:"bonk"`
	AAVE  CoingeckoPriceCore `json:"aave"`
	FLOKI CoingeckoPriceCore `json:"floki"`
	JUP   CoingeckoPriceCore `json:"jup"`
	ENS   CoingeckoPriceCore `json:"ens"`
	W     CoingeckoPriceCore `json:"w"`
	MEW   CoingeckoPriceCore `json:"mew"`
	WBTC  CoingeckoPriceCore `json:"wbtc"`
	WETH  CoingeckoPriceCore `json:"weth"`
	WSOL  CoingeckoPriceCore `json:"wsol"`
	SAND  CoingeckoPriceCore `json:"sand"`
}

type CryptoPriceMeta struct {
	UpdatedAt int64 `json:"updated_at"`
}
