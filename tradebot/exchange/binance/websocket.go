package binance

import (
	"context"
	"fmt"
	"time"
	"tradebot_go/tradebot/base"
	"tradebot_go/tradebot/core/messagebus"

	log "github.com/BitofferHub/pkg/middlewares/log"
)

// WSCliBinanceWSClientent represents a Binance WebSocket client
type BinanceWSClient struct {
	*base.WSClient
	msgBus *messagebus.MessageBus
}

// NewBinanceWSClient creates a new BinanceWSClient
func NewBinanceWSClient(accountType BinanceAccountType,
	handler base.MessageHandler, msgBus *messagebus.MessageBus) (*BinanceWSClient, error) {
	url := BinanceWebSocketURLs[accountType]

	wsClient, err := base.NewWSClient(url, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to create websocket client: %w", err)
	}

	return &BinanceWSClient{
		WSClient: wsClient,
		msgBus:   msgBus,
	}, nil
}

// Subscribe subscribes to a market data stream
func (c *BinanceWSClient) Subscribe(symbol string, streams string) error {
	c.Connect(context.Background())
	subId := fmt.Sprintf("%s@%s", symbol, streams)
	c.SubscribedStreams = append(c.SubscribedStreams, subId)
	msg := base.SubscribeMsg{
		Method: "SUBSCRIBE",
		Params: []string{subId},
		ID:     time.Now().UnixNano(),
	}

	log.Infof("Subscribing to %s@%s", symbol, streams)
	return c.WriteJSON(msg)
}

func (c *BinanceWSClient) SubscribeTrade(symbol string) error {
	return c.Subscribe(symbol, "trade")
}

func (c *BinanceWSClient) SubscribeBookL1(symbol string) error {
	return c.Subscribe(symbol, "bookTicker")
}
