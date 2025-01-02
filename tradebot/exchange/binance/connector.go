package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"tradebot_go/tradebot/core/messagebus"
)

type PublicConnector interface {
	SubscribeTrade(symbol string) error
	SubscribeBookL1(symbol string) error
}

type BinancePublicConnector struct {
	wsClient *BinanceWSClient
	msgBus   *messagebus.MessageBus
}

func NewBinancePublicConnector(msgBus *messagebus.MessageBus) (*BinancePublicConnector, error) {
	connector := &BinancePublicConnector{
		msgBus: msgBus,
	}

	wsClient, err := NewBinanceWSClient(
		BinanceAccountTypeUsdMFuturesTestnet,
		connector.HandleMessage,
		msgBus,
	)
	if err != nil {
		return nil, err
	}

	connector.wsClient = wsClient
	return connector, nil
}

func (c *BinancePublicConnector) Connect() error {
	return c.wsClient.Connect(context.Background())
}

func (c *BinancePublicConnector) Close() error {
	return c.wsClient.Close()
}

func (c *BinancePublicConnector) SubscribeTrade(symbol string) error {
	return c.wsClient.SubscribeTrade(symbol)
}

func (c *BinancePublicConnector) SubscribeBookL1(symbol string) error {
	return c.wsClient.SubscribeBookL1(symbol)
}

func (c *BinancePublicConnector) HandleMessage(msg map[string]interface{}) error {
	event := msg["e"]
	switch event {
	case "trade":
		trade, err := c.HandleTradeMessage(msg)
		if err != nil {
			return fmt.Errorf("failed to handle trade message: %v", err)
		}
		if c.msgBus != nil {
			c.msgBus.Send("trade", trade)
		}
	case "bookTicker":
		bookTicker, err := c.HandleBookL1Message(msg)
		if err != nil {
			return fmt.Errorf("failed to handle bookTicker message: %v", err)
		}
		if c.msgBus != nil {
			c.msgBus.Send("bookTicker", bookTicker)
		}
	}
	return nil
}

// HandleTradeMessage converts raw message to Trade struct
func (c *BinancePublicConnector) HandleTradeMessage(msg map[string]interface{}) (*Trade, error) {
	return parseMessage[Trade](msg)
}

// HandleBookTickerMessage converts raw message to BookTicker struct
func (c *BinancePublicConnector) HandleBookL1Message(msg map[string]interface{}) (*BookTicker, error) {
	return parseMessage[BookTicker](msg)
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
