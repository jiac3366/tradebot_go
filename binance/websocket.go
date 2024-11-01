package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	baseWSURL = "wss://stream.binance.com:9443/ws"
)

// WSClient represents a Binance WebSocket client
type WSClient struct {
	conn     *websocket.Conn
	mu       sync.Mutex
	handlers map[string]MessageHandler
	done     chan struct{}
	closed   bool
}

// MessageHandler is a function type for handling different message types
type MessageHandler func([]byte) error

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

// NewWSClient creates a new WebSocket client
func NewWSClient() (*WSClient, error) {
	c := &WSClient{
		handlers: make(map[string]MessageHandler),
		done:     make(chan struct{}),
	}
	return c, nil
}

// Connect establishes a WebSocket connection
func (c *WSClient) Connect(ctx context.Context) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, baseWSURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to websocket: %w", err)
	}

	c.conn = conn

	// Start message handling loop
	go c.messageLoop()

	// Start ping/pong handler
	go c.keepAlive()

	return nil
}

// Subscribe subscribes to a market data stream
func (c *WSClient) Subscribe(symbol string, streams []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	msg := SubscribeMsg{
		Method: "SUBSCRIBE",
		Params: make([]string, len(streams)),
		ID:     time.Now().UnixNano(),
	}

	// Format stream names
	for i, stream := range streams {
		msg.Params[i] = fmt.Sprintf("%s@%s", symbol, stream)
	}

	return c.writeJSON(msg)
}

// RegisterHandler registers a message handler for a specific stream
func (c *WSClient) RegisterHandler(stream string, handler MessageHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[stream] = handler
}

// messageLoop handles incoming WebSocket messages
func (c *WSClient) messageLoop() {
	for {
		select {
		case <-c.done:
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("error reading message: %v", err)
				return
			}

			// Handle the message
			if err := c.handleMessage(message); err != nil {
				log.Printf("error handling message: %v", err)
			}
		}
	}
}

// handleMessage processes incoming messages
func (c *WSClient) handleMessage(message []byte) error {
	// Parse message to determine type and route to appropriate handler
	var raw map[string]interface{}
	if err := json.Unmarshal(message, &raw); err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// Check if it's a trade message
	if eventType, ok := raw["e"].(string); ok {
		switch eventType {
		case "trade":
			if handler, exists := c.handlers["trade"]; exists {
				return handler(message)
			}
		}
	}

	return nil
}

// keepAlive handles ping/pong messages
func (c *WSClient) keepAlive() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PongMessage, nil); err != nil {
				log.Printf("failed to send pong: %v", err)
				return
			}
		}
	}
}

// writeJSON sends a JSON message through the WebSocket connection
func (c *WSClient) writeJSON(v interface{}) error {
	return c.conn.WriteJSON(v)
}

// Close closes the WebSocket connection
func (c *WSClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If already closed, return immediately
	if c.closed {
		return nil
	}

	c.closed = true
	close(c.done)
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
