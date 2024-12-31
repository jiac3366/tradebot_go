package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	log "github.com/BitofferHub/pkg/middlewares/log"
)

// Add new connection states
const (
	StatusDisconnected = "disconnected"
	StatusConnected    = "connected"
	StatusReconnecting = "reconnecting"
)

type MessageHandler func(message map[string]interface{}) error

// WSClient represents a WebSocket client
type WSClient struct {
	url        string
	handler    MessageHandler
	done       chan struct{}
	status     string
	mu         sync.RWMutex
	conn       *websocket.Conn
	maxRetries int

	// Add reconnection configuration
	reconnectWait    time.Duration
	reconnectAttempt int
}

// NewWSClient creates a new WebSocket client with improved configuration
func NewWSClient(url string, handler MessageHandler) (*WSClient, error) {
	if url == "" {
		return nil, fmt.Errorf("websocket URL cannot be empty")
	}
	if handler == nil {
		return nil, fmt.Errorf("message handler cannot be nil")
	}

	return &WSClient{
		url:           url,
		handler:       handler,
		done:          make(chan struct{}),
		status:        StatusDisconnected,
		maxRetries:    5,
		reconnectWait: 5 * time.Second,
	}, nil
}

// Connect with improved error handling and connection management
func (c *WSClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	if c.status == StatusConnected {
		c.mu.Unlock()
		return nil
	}
	c.status = StatusReconnecting
	c.mu.Unlock()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.url, nil)
	if err != nil {
		c.mu.Lock()
		c.status = StatusDisconnected
		c.mu.Unlock()
		return fmt.Errorf("websocket connection failed: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.status = StatusConnected
	c.reconnectAttempt = 0
	c.mu.Unlock()

	// Start handlers in separate goroutines
	go c.messageLoop(ctx)
	go c.Ping(1 * time.Second)

	log.Infof("Successfully connected to WebSocket at %s", c.url)
	return nil
}

// messageLoop with improved error handling and reconnection logic
func (c *WSClient) messageLoop(ctx context.Context) {
	defer func() {
		c.mu.Lock()
		if c.conn != nil {
			c.conn.Close()
		}
		c.status = StatusDisconnected
		c.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.done:
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err) {
					log.Errorf("Unexpected websocket closure: %v", err)
					c.tryReconnect(ctx)
					return
				}
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					log.Errorf("Websocket read error: %v", err)
					c.tryReconnect(ctx)
					return
				}
				return
			}

			if err := c.handleMessage(message); err != nil {
				log.Errorf("Message handling error: %v", err)
			}
		}
	}
}

// tryReconnect implements exponential backoff reconnection
func (c *WSClient) tryReconnect(ctx context.Context) {
	c.mu.Lock()
	if c.status == StatusReconnecting {
		c.mu.Unlock()
		return
	}
	c.status = StatusReconnecting
	c.mu.Unlock()

	for c.reconnectAttempt < c.maxRetries {
		wait := c.reconnectWait * time.Duration(1<<uint(c.reconnectAttempt))
		log.Infof("Attempting to reconnect in %v (attempt %d/%d)", wait, c.reconnectAttempt+1, c.maxRetries)

		time.Sleep(wait)

		if err := c.Connect(ctx); err == nil {
			return
		}

		c.mu.Lock()
		c.reconnectAttempt++
		c.mu.Unlock()
	}

	log.Error("Max reconnection attempts reached")
}

// handleMessage processes incoming messages
func (c *WSClient) handleMessage(message []byte) error {
	// Parse the message
	var raw map[string]interface{}
	if err := json.Unmarshal(message, &raw); err != nil {
		// Handle ping frame
		if string(message) == "ping" {
			if err := c.conn.WriteMessage(websocket.TextMessage, []byte("pong")); err != nil {
				return fmt.Errorf("failed to send pong: %w", err)
			}
			log.Debug("Sent pong response")
			return nil
		}
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// Handle subscription response
	if _, ok := raw["result"]; ok {
		log.Info("Received subscription response")
		return nil
	}

	// Handle the message using the provided handler
	return c.handler(raw)
}

// WriteJSON sends a JSON message through the WebSocket connection
func (c *WSClient) WriteJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("connection is not established")
	}

	return c.conn.WriteJSON(v)
}

// Close closes the WebSocket connection
func (c *WSClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.status == StatusDisconnected {
		return nil
	}
	c.status = StatusDisconnected

	close(c.done)
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Ping sends periodic ping messages to keep the connection alive
func (c *WSClient) Ping(interval time.Duration) {
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
			c.mu.RLock()
			if c.conn == nil || c.status != StatusConnected {
				c.mu.RUnlock()
				log.Debug("Connection not established, skipping ping")
				continue
			}
			c.mu.RUnlock()

			c.mu.Lock()
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Errorf("Failed to write ping message: %v", err)
				c.mu.Unlock()
				c.tryReconnect(context.Background())
				continue
			}
			c.mu.Unlock()

			log.Debug("Ping message sent successfully")
		}
	}
}
