package binance

// SubscribeMsg represents a subscription message
type SubscribeMsg struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	ID     int64    `json:"id"`
}

// Trade represents a trade message from Binance
type Trade struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	TradeID   int64  `json:"t"`
	Price     string `json:"p"`
	Quantity  string `json:"q"`
	TradeTime int64  `json:"T"`
	IsMaker   bool   `json:"m"`
	Ignore    bool   `json:"M"`
}

type BinanceAccountType string

const (
	BinanceAccountTypeSpot                BinanceAccountType = "SPOT"
	BinanceAccountTypeMargin              BinanceAccountType = "MARGIN"
	BinanceAccountTypeIsolatedMargin      BinanceAccountType = "ISOLATED_MARGIN"
	BinanceAccountTypeUsdMFutures         BinanceAccountType = "USD_M_FUTURE"
	BinanceAccountTypeCoinMFutures        BinanceAccountType = "COIN_M_FUTURE"
	BinanceAccountTypePortfolioMargin     BinanceAccountType = "PORTFOLIO_MARGIN"
	BinanceAccountTypeSpotTestnet         BinanceAccountType = "SPOT_TESTNET"
	BinanceAccountTypeUsdMFuturesTestnet  BinanceAccountType = "USD_M_FUTURE_TESTNET"
	BinanceAccountTypeCoinMFuturesTestnet BinanceAccountType = "COIN_M_FUTURE_TESTNET"
)

// WebSocketURLs maps account types to their WebSocket endpoints
var BinanceWebSocketURLs = map[BinanceAccountType]string{
	BinanceAccountTypeSpot:                "wss://stream.binance.com:9443/ws",
	BinanceAccountTypeMargin:              "wss://stream.binance.com:9443/ws",
	BinanceAccountTypeIsolatedMargin:      "wss://stream.binance.com:9443/ws",
	BinanceAccountTypeUsdMFutures:         "wss://fstream.binance.com/ws",
	BinanceAccountTypeCoinMFutures:        "wss://dstream.binance.com/ws",
	BinanceAccountTypePortfolioMargin:     "wss://fstream.binance.com/pm/ws",
	BinanceAccountTypeSpotTestnet:         "wss://testnet.binance.vision/ws",
	BinanceAccountTypeUsdMFuturesTestnet:  "wss://stream.binancefuture.com/ws",
	BinanceAccountTypeCoinMFuturesTestnet: "wss://dstream.binancefuture.com/ws",
}
