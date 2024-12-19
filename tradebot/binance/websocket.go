package binance

import (
	"fmt"
	"time"

	"tradebot_go/tradebot/core"
)

const (
	baseWSURL = "wss://stream.binance.com:9443/ws"
)

// WSCliBinanceWSClientent represents a Binance WebSocket client
type BinanceWSClient struct {
	*core.WSClient // core.WSClient is wrong as we're copying a mutex when embedding the WSClient.
}

// NewBinanceWSClient creates a new BinanceWSClient
func NewBinanceWSClient(accountType BinanceAccountType, handler core.MessageHandler) (*BinanceWSClient, error) {

	client := &BinanceWSClient{}

	url := fmt.Sprintf("%s/%s", BinanceWebSocketURLs[accountType], "stream")

	wsClient, err := core.NewWSClient(url, client.WsHandleData)
	if err != nil {
		return nil, err
	}
	return &BinanceWSClient{
		WSClient: wsClient,
	}, nil
}

// Subscribe subscribes to a market data stream
func (c *BinanceWSClient) Subscribe(symbol string, streams []string) error {

	msg := SubscribeMsg{
		Method: "SUBSCRIBE",
		Params: make([]string, len(streams)),
		ID:     time.Now().UnixNano(),
	}

	// Format stream names
	for i, stream := range streams {
		msg.Params[i] = fmt.Sprintf("%s@%s", symbol, stream)
	}

	return c.WriteJSON(msg)
}

func (c *BinanceWSClient) WsHandleData(raw map[string]interface{}) error {
	return nil
}
