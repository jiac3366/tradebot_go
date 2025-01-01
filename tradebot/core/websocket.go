package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	log "github.com/BitofferHub/pkg/middlewares/log"
	"golang.org/x/time/rate"
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
	reconnectWait     time.Duration
	reconnectAttempt  int
	SubscribedStreams []string
	limiter           *rate.Limiter // Rate limiter for resubscription
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
		limiter:       rate.NewLimiter(rate.Every(300*time.Millisecond), 1), // 300ms per request, burst size of 1
	}, nil
}

func (c *WSClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status == StatusConnected
}

// Connect with improved error handling and connection management
func (c *WSClient) Connect(ctx context.Context) error {
	defer func() {
		c.mu.Lock()
		if c.status != StatusConnected {
			c.status = StatusDisconnected
		}
		c.mu.Unlock()
	}()

	if c.IsConnected() {
		return nil
	}
	c.mu.Lock()
	c.status = StatusReconnecting
	c.mu.Unlock()

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, c.url, nil)
	if err != nil {
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
	conn.SetReadDeadline(time.Time{})
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(15 * time.Minute))
		return nil
	})

	c.mu.Lock()
	c.conn = conn
	c.status = StatusConnected
	c.reconnectAttempt = 0
	c.mu.Unlock()

	go c.messageLoop(ctx)

	// 不需要主动发送ping，因为Binance服务器会发送ping
	// go c.Ping(30 * time.Second)

	log.Infof("Successfully connected to WebSocket at %s", c.url)
	return nil
}

// messageLoop with improved error handling and reconnection logic
func (c *WSClient) messageLoop(ctx context.Context) {

	// 设置一个更合理的 read deadline
	const readTimeout = 1 * time.Minute

	for {
		select {
		case <-ctx.Done():
			log.Infof("messageLoop: ctx.Done")
			return
		case <-c.done:
			log.Infof("messageLoop: done")
			return
		default:
			// 为每次读取设置新的 deadline
			c.conn.SetReadDeadline(time.Now().Add(readTimeout))

			messageType, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Errorf("ReadMessage error: %v, messageType: %d", err, messageType)
				c.tryReconnect()
				return
			}

			if err := c.handleMessage(message); err != nil {
				log.Errorf("Message handling error: %v", err)
			}
		}
	}
}

// tryReconnect implements exponential backoff reconnection
func (c *WSClient) tryReconnect() {
	defer func() {
		c.mu.Lock()
		if c.status != StatusConnected {
			c.status = StatusDisconnected
		}
		c.mu.Unlock()
	}()

	log.Infof("tryReconnect")

	c.mu.Lock()
	if c.status == StatusReconnecting {
		c.mu.Unlock()
		return
	}

	if c.conn != nil {
		log.Infof("tryReconnect: Closing websocket connection first")
		c.conn.Close()
		c.status = StatusDisconnected
	}
	c.mu.Unlock()

	for c.reconnectAttempt < c.maxRetries {
		wait := c.reconnectWait * time.Duration(1<<uint(c.reconnectAttempt))
		log.Infof("Attempting to reconnect in %v (attempt %d/%d)", wait, c.reconnectAttempt+1, c.maxRetries)

		time.Sleep(wait)

		if err := c.Connect(context.Background()); err == nil {
			log.Info("Reconnected successfully")
			break
		}

		c.mu.Lock()
		c.reconnectAttempt++
		c.mu.Unlock()
	}

	if c.reconnectAttempt == c.maxRetries {
		log.Error("Max reconnection attempts reached")
	} else {
		// Resubscribe with rate limiting
		log.Infof("Resubscribing with rate limiting %v", c.SubscribedStreams)
		for _, subId := range c.SubscribedStreams {
			// Wait for rate limiter
			err := c.limiter.Wait(context.Background())
			if err != nil {
				log.Errorf("Rate limiter error: %v", err)
				continue
			}

			err = c.WriteJSON(SubscribeMsg{
				Method: "SUBSCRIBE",
				Params: []string{subId},
				ID:     time.Now().UnixNano(),
			})
			if err != nil {
				log.Errorf("Failed to resubscribe to %s: %v", subId, err)
			} else {
				log.Infof("Resubscribed to %s", subId)
			}
		}
	}
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

func (c *WSClient) Close2() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// WriteJson --> writejson
