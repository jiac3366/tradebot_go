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

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, c.url, nil)
	if err != nil {
		c.mu.Lock()
		c.status = StatusDisconnected
		c.mu.Unlock()
		return fmt.Errorf("websocket connection failed: %w", err)
	}

	// 设置 Binance ping-pong 处理
	// Binance 服务器每3分钟发送一次 ping frame，我们需要在10分钟内回应 pong frame
	conn.SetPingHandler(func(appData string) error {
		err := conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(10*time.Second))
		if err != nil {
			log.Errorf("Failed to send pong: %v", err)
			return err
		}
		log.Debug("Sent pong response")
		return nil
	})

	// 设置读取超时为15分钟（大于Binance的10分钟限制）
	conn.SetReadDeadline(time.Now().Add(15 * time.Minute))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(15 * time.Minute))
		return nil
	})

	c.mu.Lock()
	c.conn = conn
	c.status = StatusConnected
	c.reconnectAttempt = 0
	c.mu.Unlock()

	// Start message loop
	go c.messageLoop(ctx)

	// 不需要主动发送ping，因为Binance服务器会发送ping
	// go c.Ping(30 * time.Second)

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
			messageType, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Errorf("ReadMessage error: %v, messageType: %d", err, messageType)
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
	log.Debugf("Received message: %s", string(message))

	var raw map[string]interface{}
	if err := json.Unmarshal(message, &raw); err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// Handle subscription response
	if _, ok := raw["result"]; ok {
		return nil
	}

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
