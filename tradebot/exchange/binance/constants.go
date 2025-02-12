package binance

// Trade represents a trade message from Binance
//
//	{
//		"e": "trade",       // Event type
//		"E": 1672515782136, // Event time
//		"s": "BNBBTC",      // Symbol
//		"t": 12345,         // Trade ID
//		"p": "0.001",       // Price
//		"q": "100",         // Quantity
//		"T": 1672515782136, // Trade time
//		"m": true,          // Is the buyer the market maker?
//		"M": true           // Ignore
//	}
type Trade struct {
    EventType  string `json:"e" validate:"required"`
    EventTime  int64  `json:"E" validate:"required"`
    Symbol     string `json:"s" validate:"required"`
    TradeID    int64  `json:"t" validate:"required"`
    Price      string `json:"p" validate:"required"`
    Quantity   string `json:"q" validate:"required"`
    TradeTime  int64  `json:"T" validate:"required"`
    IsMaker    bool   `json:"m" validate:"required"`
    Ignore     bool   `json:"M"`
    MarketType string `json:"X"`
}

// bookTicker
//
//	{
//		"u":400900217,     // order book updateId
//		"s":"BNBUSDT",     // symbol
//		"b":"25.35190000", // best bid price
//		"B":"31.21000000", // best bid qty
//		"a":"25.36520000", // best ask price
//		"A":"40.66000000"  // best ask qty
//	}
type BookTicker struct {
	UpdateID int64  `json:"u"`
	Symbol   string `json:"s"`
	BidPrice string `json:"b"`
	BidQty   string `json:"B"`
	AskPrice string `json:"a"`
	AskQty   string `json:"A"`
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


var BinanceHttpURLs = map[BinanceAccountType]string{
	BinanceAccountTypeSpot:                "https://api.binance.com",
	BinanceAccountTypeMargin:              "https://api.binance.com",
	BinanceAccountTypeIsolatedMargin:      "https://api.binance.com",
	BinanceAccountTypeUsdMFutures:         "https://fapi.binance.com",
	BinanceAccountTypeCoinMFutures:        "https://dapi.binance.com",
	BinanceAccountTypePortfolioMargin:     "https://fapi.binance.com",
	BinanceAccountTypeSpotTestnet:         "https://testnet.binance.vision",
	BinanceAccountTypeUsdMFuturesTestnet:  "https://testnet.binancefuture.com",
	BinanceAccountTypeCoinMFuturesTestnet: "https://testnet.dapi.binance.com",
}
