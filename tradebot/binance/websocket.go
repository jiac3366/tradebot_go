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

// Connect establishes the websocket connection
func (c *BinanceWSClient) Connect(ctx context.Context) error {
	return c.wsClient.Connect(ctx)
}

// HandleTradeMessage converts raw message to Trade struct
func (c *BinanceWSClient) HandleTradeMessage(msg map[string]interface{}) (*Trade, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	trade := &Trade{}
	if err := json.Unmarshal(jsonBytes, trade); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trade: %w", err)
	}

	return trade, nil
}
