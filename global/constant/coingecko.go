package constant

var (
	CoingeckoAPI      = "https://api.coingecko.com/api/v3"
	CoingeckoGetPrice = CoingeckoAPI + "/simple/price"
)

// Redis keys for cached market prices.
const (
	CRYPTO_PRICE_RAW    = "CRYPTO_PRICE_RAW"    // CoinGecko id-keyed JSON (bitcoin/ethereum/...)
	CRYPTO_PRICE_SYMBOL = "CRYPTO_PRICE_SYMBOL" // Symbol-keyed JSON (btc/eth/...) for deshop
	CRYPTO_PRICE_META   = "CRYPTO_PRICE_META"   // {"updated_at": unix_ms}
)

var (
	IDS_BTC   = "bitcoin"
	IDS_ETH   = "ethereum"
	IDS_USDT  = "tether"
	IDS_BNB   = "binancecoin"
	IDS_SOL   = "solana"
	IDS_USDC  = "usd-coin"
	IDS_XRP   = "ripple"
	IDS_TON   = "the-open-network"
	IDS_DOGE  = "dogecoin"
	IDS_ADA   = "cardano"
	IDS_TRX   = "tron"
	IDS_AVAX  = "avalanche-2"
	IDS_SHIB  = "shiba-inu"
	IDS_DOT   = "polkadot"
	IDS_LINK  = "chainlink"
	IDS_BCH   = "bitcoin-cash"
	IDS_DAI   = "dai"
	IDS_LTC   = "litecoin"
	IDS_POL   = "polygon-ecosystem-token"
	IDS_UNI   = "uniswap"
	IDS_PEPE  = "pepe"
	IDS_FIL   = "filecoin"
	IDS_ARB   = "arbitrum"
	IDS_OP    = "optimism"
	IDS_WIF   = "dogwifcoin"
	IDS_NOT   = "notcoin"
	IDS_BONK  = "bonk"
	IDS_AAVE  = "aave"
	IDS_FLOKI = "floki"
	IDS_JUP   = "jupiter"
	IDS_ENS   = "ethereum-name-service"
	IDS_W     = "wormhole"
	IDS_MEW   = "meow"
	IDS_WBTC  = "wrapped-bitcoin"
	IDS_WETH  = "weth"
	IDS_WSOL  = "wrapped-solana"
	IDS_SAND  = "the-sandbox"
)

var COIN_IDS = []string{
	IDS_BTC, IDS_ETH, IDS_USDT, IDS_BNB, IDS_SOL, IDS_USDC, IDS_XRP, IDS_TON,
	IDS_DOGE, IDS_ADA, IDS_TRX, IDS_AVAX, IDS_SHIB, IDS_DOT, IDS_LINK, IDS_BCH,
	IDS_DAI, IDS_LTC, IDS_POL, IDS_UNI, IDS_PEPE, IDS_FIL, IDS_ARB, IDS_OP,
	IDS_WIF, IDS_NOT, IDS_BONK, IDS_AAVE, IDS_FLOKI, IDS_JUP, IDS_ENS, IDS_W,
	IDS_MEW, IDS_WBTC, IDS_WETH, IDS_WSOL, IDS_SAND,
}
