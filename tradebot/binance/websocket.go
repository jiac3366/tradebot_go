package binance

import (
	"fmt"
	"time"
	"tradebot_go/tradebot/core"
)

// WSCliBinanceWSClientent represents a Binance WebSocket client
type BinanceWSClient struct {
	*core.WSClient // core.WSClient is wrong as we're copying a mutex when embedding the WSClient.
}

// NewBinanceWSClient creates a new BinanceWSClient
func NewBinanceWSClient(accountType BinanceAccountType, handler core.MessageHandler) (*BinanceWSClient, error) {

	// client := &BinanceWSClient{}
	url := BinanceWebSocketURLs[accountType]

	wsClient, err := core.NewWSClient(url, handler)
	if err != nil {
		return nil, err
	}
	return &BinanceWSClient{
		WSClient: wsClient,
	}, nil
}

// Subscribe subscribes to a market data stream
func (c *BinanceWSClient) Subscribe(symbol string, streams string) error {

	msg := SubscribeMsg{
		Method: "SUBSCRIBE",
		Params: []string{fmt.Sprintf("%s@%s", symbol, streams)},
		ID:     time.Now().UnixNano(),
	}

	return c.WriteJSON(msg)
}

// HandleTradeMessage converts raw message to Trade struct
func (c *BinanceWSClient) HandleTradeMessage(msg map[string]interface{}) (*Trade, error) {
	// 将map数据转换为Trade结构体
	trade := &Trade{}

	// 使用类型断言获取数据
	if e, ok := msg["e"].(string); ok {
		trade.EventType = e
	}
	if t, ok := msg["E"].(float64); ok {
		trade.EventTime = int64(t)
	}
	if s, ok := msg["s"].(string); ok {
		trade.Symbol = s
	}
	if tid, ok := msg["t"].(float64); ok {
		trade.TradeID = int64(tid)
	}
	if p, ok := msg["p"].(string); ok {
		trade.Price = p
	}
	if q, ok := msg["q"].(string); ok {
		trade.Quantity = q
	}

	if tt, ok := msg["T"].(float64); ok {
		trade.TradeTime = int64(tt)
	}
	if m, ok := msg["m"].(bool); ok {
		trade.IsMaker = m
	}
	if x, ok := msg["X"].(bool); ok {
		trade.MarketType = x
	}
	return trade, nil
}
