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
	wsClient *base.WSClient // Change from embedding to composition
	msgBus   *messagebus.MessageBus
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
		wsClient: wsClient,
		msgBus:   msgBus,
	}, nil
}

// Subscribe subscribes to a market data stream
func (c *BinanceWSClient) Subscribe(symbol string, streams string) error {
	c.Connect(context.Background())
	subId := fmt.Sprintf("%s@%s", symbol, streams)
	c.wsClient.SubscribedStreams = append(c.wsClient.SubscribedStreams, subId)
	msg := base.SubscribeMsg{
		Method: "SUBSCRIBE",
		Params: []string{subId},
		ID:     time.Now().UnixNano(),
	}

	log.Infof("Subscribing to %s@%s", symbol, streams)
	return c.wsClient.WriteJSON(msg)
}

// Close closes the websocket connection
func (c *BinanceWSClient) Close() error {
	return c.wsClient.Close()
}

// for testing reconnection function
func (c *BinanceWSClient) CloseConnection() error {
	return c.wsClient.CloseConnection()
}

// Connect establishes the websocket connection
func (c *BinanceWSClient) Connect(ctx context.Context) error {
	return c.wsClient.Connect(ctx)
}

func (c *BinanceWSClient) SubscribeTrade(symbol string) error {
	return c.Subscribe(symbol, "trade")
}

func (c *BinanceWSClient) SubscribeBookL1(symbol string) error {
	return c.Subscribe(symbol, "bookTicker")
}
