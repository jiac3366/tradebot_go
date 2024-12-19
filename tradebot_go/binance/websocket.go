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

	status *WSClientStatus
}

type WSClientStatus struct {
	closed    bool
	connected bool
}

// MessageHandler is a function type for handling different message types
type MessageHandler func([]byte) error

// Trade represents a trade message from Binance

// NewWSClient creates a new WebSocket client
func NewWSClient() (*WSClient, error) {
	c := &WSClient{
		handlers: make(map[string]MessageHandler),
		done:     make(chan struct{}),

		status: &WSClientStatus{
			closed:    true,
			connected: false,
		},
	}
	return c, nil
}

func (c *WSClient) isConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status.connected
}

func (c *WSClient) isClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status.closed
}

func (c *WSClient) setConnected(connected bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status.connected = connected
}

func (c *WSClient) setClosed(closed bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status.closed = closed
}

// Connect establishes a WebSocket connection
func (c *WSClient) Connect(ctx context.Context) error {
	if c.isConnected() {
		return nil
	}
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

	// connection break resubs

	// Start ping/pong handler
	go c.ping()

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

// ping/pong messages
func (c *WSClient) ping() {
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
