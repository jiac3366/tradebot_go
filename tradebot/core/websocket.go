package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

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

	conn, _, err := dialer.DialContext(ctx, c.url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to websocket: %w", err)
	}
	c.conn = conn
	c.setConnected()
	fmt.Println("connected")

	go c.messageLoop()
	// go c.Ping(1 * time.Second)

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
				logger.Printf("error reading message: %v", err)
				return
			}

			// Handle the message
			if err := c.handleMessage(message); err != nil {
				logger.Printf("error handling message: %v", err)
			}
		}
	}
}

// handleMessage processes incoming messages
func (c *WSClient) handleMessage(message []byte) error {
	// 1. 首先检查是否是 ping frame
	var raw map[string]interface{}
	if err := json.Unmarshal(message, &raw); err != nil {
		// 可能是 ping frame，需要回复 pong
		if string(message) == "ping" {
			if err := c.conn.WriteMessage(websocket.TextMessage, []byte("pong")); err != nil {
				return fmt.Errorf("failed to send pong: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// 2. 检查是否是订阅响应消息
	if _, ok := raw["result"]; ok {
		// 这是订阅响应消息，可以记录日志但不需要进一步处理
		logger.Printf("Subscription response: %+v", raw)
		return nil
	}

	fmt.Printf("%+v\n", raw)
	return c.handler(raw)
}

// writeJSON sends a JSON message through the WebSocket connection
func (c *WSClient) WriteJSON(v interface{}) error {
	// fmt v
	fmt.Println(v)
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

func (c *WSClient) Ping(interval time.Duration) {
	// default to 3 minutes
	if interval == 0 {
		interval = 3 * time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.mu.Lock()
			if c.conn == nil || c.status != "connected" {
				c.mu.Unlock()
				logger.Printf("connection not established, skipping ping")
				return
			}

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.mu.Unlock()
				logger.Printf("failed to send ping: %v", err)
				return
			}
			c.mu.Unlock()
		}
	}
}
