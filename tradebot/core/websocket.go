package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSClient represents a Binance WebSocket client
type WSClient struct {
	url     string
	handler MessageHandler
	done    chan struct{}
	status  string

	mu   sync.Mutex
	conn *websocket.Conn
}

// MessageHandler is a function type for handling different message types
type MessageHandler func(map[string]interface{}) error

// NewWSClient creates a new WebSocket client
func NewWSClient(url string, handler MessageHandler) (*WSClient, error) {
	c := &WSClient{
		url:     url,
		handler: handler,
		done:    make(chan struct{}),
		status:  "disconnected",
	}
	return c, nil
}

func (c *WSClient) isConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status == "connected"
}

func (c *WSClient) setConnected() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status = "connected"
}

// Connect establishes a WebSocket connection
func (c *WSClient) Connect(ctx context.Context) error {
	if c.isConnected() {
		return nil
	}
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to websocket: %w", err)
	}
	c.conn = conn

	go c.messageLoop()
	go c.ping()

	return nil
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
	// todo: map[string]interface{}
	var raw map[string]interface{}
	if err := json.Unmarshal(message, &raw); err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	return c.WsHandleData(raw)
}

// handle by exchange stucture
func (c *WSClient) WsHandleData(raw map[string]interface{}) error {
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
func (c *WSClient) WriteJSON(v interface{}) error {
	return c.conn.WriteJSON(v)
}

// Close closes the WebSocket connection
func (c *WSClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.status == "disconnected" {
		return nil
	}
	c.status = "disconnected"

	close(c.done)
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// reconnect
func (c *WSClient) reconnect() error {
	return c.Connect(context.Background())
}
