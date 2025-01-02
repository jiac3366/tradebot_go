package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"tradebot_go/tradebot/core"

	log "github.com/BitofferHub/pkg/middlewares/log"
)

// WSCliBinanceWSClientent represents a Binance WebSocket client
type BinanceWSClient struct {
	wsClient *core.WSClient // Change from embedding to composition
}

// NewBinanceWSClient creates a new BinanceWSClient
func NewBinanceWSClient(accountType BinanceAccountType, handler core.MessageHandler) (*BinanceWSClient, error) {
	url := BinanceWebSocketURLs[accountType]
	wsClient, err := core.NewWSClient(url, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to create websocket client: %w", err)
	}

	return &BinanceWSClient{
		wsClient: wsClient,
	}, nil
}

// Subscribe subscribes to a market data stream
func (c *BinanceWSClient) Subscribe(symbol string, streams string) error {
	c.Connect(context.Background())
	subId := fmt.Sprintf("%s@%s", symbol, streams)
	c.wsClient.SubscribedStreams = append(c.wsClient.SubscribedStreams, subId)
	msg := core.SubscribeMsg{
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

// parseMessage is a generic function to handle websocket messages
func parseMessage[T any](msg map[string]interface{}) (*T, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	var result T
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &result, nil
}

// HandleTradeMessage converts raw message to Trade struct
func (c *BinanceWSClient) HandleTradeMessage(msg map[string]interface{}) (*Trade, error) {
	return parseMessage[Trade](msg)
}

// HandleBookTickerMessage converts raw message to BookTicker struct
func (c *BinanceWSClient) HandleBookL1Message(msg map[string]interface{}) (*BookTicker, error) {
	return parseMessage[BookTicker](msg)
}

func (c *BinanceWSClient) SubscribeTrade(symbol string) error {
	return c.Subscribe(symbol, "trade")
}

func (c *BinanceWSClient) SubscribeBookL1(symbol string) error {
	return c.Subscribe(symbol, "bookTicker")
}

// func (c *BinanceWSClient) HandleMessage(msg map[string]interface{}) {
// 	event := msg["e"]
// 	switch event {
// 	case "trade":
// 		trade, err := c.HandleTradeMessage(msg)
// 		if err != nil {
// 			// ignore error
// 			log.Errorf("failed to handle trade message: %v", err)
// 		}
// 		// todo: send to msgbus
// 	case "bookTicker":
// 		bookTicker, err := c.HandleBookL1Message(msg)
// 		if err != nil {
// 			// ignore error
// 			log.Errorf("failed to handle bookTicker message: %v", err)
// 		}
// 		// todo: send to msgbus
// 	}
// }
