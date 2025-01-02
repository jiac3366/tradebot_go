package messagebus

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/google/uuid"
)

// MessageBus represents a generic message bus implementation
type MessageBus struct {
	traderID   string
	instanceID uuid.UUID
	name       string

	// Concurrent maps for thread safety
	endpoints sync.Map // map[string]Handler
	patterns  sync.Map // map[string][]Subscription
	subs      sync.Map // map[Subscription][]string
	corrIndex sync.Map // map[uuid.UUID]Handler

	// Statistics
	sentCount int64
	reqCount  int64
	resCount  int64
	pubCount  int64

	logger *log.Logger
	config *Config
	mu     sync.RWMutex

	subscriptions sync.Map // map[SubscriptionID]Subscription
	topicSubs     sync.Map // map[string][]SubscriptionID
}

// NewMessageBus creates a new message bus instance
func NewMessageBus(
	traderID string,
	instanceID uuid.UUID,
	name string,
	config *Config,
) *MessageBus {
	if name == "" {
		name = "MessageBus"
	}

	if config == nil {
		config = &Config{
			BufferIntervalMS: 100,
			AutoTrimMins:     60,
		}
	}

	mb := &MessageBus{
		traderID:   traderID,
		instanceID: instanceID,
		name:       name,
		config:     config,
		logger:     log.New(log.Writer(), fmt.Sprintf("[%s] ", name), log.LstdFlags),
	}

	// Validate config
	if config.BufferIntervalMS > 1000 {
		mb.logger.Printf("Warning: High buffer_interval_ms at %d, recommended range is [10, 1000] milliseconds",
			config.BufferIntervalMS)
	}

	// Log configuration
	mb.logger.Printf("Configuration: %+v", config)

	return mb
}

// Register registers a handler for an endpoint
func (mb *MessageBus) Register(endpoint string, handler Handler) error {
	if endpoint == "" {
		return fmt.Errorf("endpoint cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	if _, loaded := mb.endpoints.LoadOrStore(endpoint, handler); loaded {
		return fmt.Errorf("endpoint %s already registered", endpoint)
	}

	mb.logger.Printf("Added endpoint '%s'", endpoint)
	return nil
}

// Deregister removes a handler for an endpoint
func (mb *MessageBus) Deregister(endpoint string, handler Handler) error {
	if endpoint == "" {
		return fmt.Errorf("endpoint cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	h, exists := mb.endpoints.LoadAndDelete(endpoint)
	if !exists {
		return fmt.Errorf("endpoint %s not registered", endpoint)
	}

	// Convert handlers to comparable values using reflect
	if reflect.ValueOf(h).Pointer() != reflect.ValueOf(handler).Pointer() {
		// Restore the handler since it didn't match
		mb.endpoints.Store(endpoint, h)
		return fmt.Errorf("handler mismatch for endpoint %s", endpoint)
	}

	mb.logger.Printf("Removed endpoint '%s'", endpoint)
	return nil
}

// SubscriptionID represents a unique identifier for a subscription
type SubscriptionID string

// generateSubscriptionID generates a new unique subscription ID
func generateSubscriptionID() SubscriptionID {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return SubscriptionID(hex.EncodeToString(bytes))
}

// // Subscribe subscribes to a topic with a handler and priority
// func (mb *MessageBus) Subscribe(topic string, handler Handler, priority int) error {
// 	if topic == "" {
// 		return fmt.Errorf("topic cannot be empty")
// 	}
// 	if handler == nil {
// 		return fmt.Errorf("handler cannot be nil")
// 	}
// 	if priority < 0 {
// 		return fmt.Errorf("priority cannot be negative")
// 	}

// 	subID := generateSubscriptionID()
// 	sub := Subscription{
// 		ID:       subID,
// 		Topic:    topic,
// 		Handler:  handler,
// 		Priority: priority,
// 	}

// 	// Store the subscription
// 	mb.subscriptions.Store(subID, sub)

// 	// Update topic subscriptions
// 	var subs []SubscriptionID
// 	if value, exists := mb.topicSubs.Load(topic); exists {
// 		subs = value.([]SubscriptionID)
// 	}
// 	subs = append(subs, subID)
// 	mb.topicSubs.Store(topic, subs)

// 	mb.logger.Printf("Added subscription: %v for topic: %s", subID, topic)
// 	return nil
// }

// // Unsubscribe removes a subscription
// func (mb *MessageBus) Unsubscribe(topic string, handler Handler) error {
// 	if topic == "" {
// 		return fmt.Errorf("topic cannot be empty")
// 	}
// 	if handler == nil {
// 		return fmt.Errorf("handler cannot be nil")
// 	}

// 	if value, exists := mb.topicSubs.Load(topic); exists {
// 		subs := value.([]SubscriptionID)
// 		var newSubs []SubscriptionID

// 		for _, subID := range subs {
// 			if sub, ok := mb.subscriptions.Load(subID); ok {
// 				subscription := sub.(Subscription)
// 				// Compare handlers using reflect
// 				if reflect.ValueOf(subscription.Handler).Pointer() != reflect.ValueOf(handler).Pointer() {
// 					newSubs = append(newSubs, subID)
// 				} else {
// 					mb.subscriptions.Delete(subID)
// 				}
// 			}
// 		}

// 		if len(newSubs) == 0 {
// 			mb.topicSubs.Delete(topic)
// 		} else {
// 			mb.topicSubs.Store(topic, newSubs)
// 		}
// 	}

// 	mb.logger.Printf("Removed subscription for topic: %s", topic)
// 	return nil
// }

// // Publish publishes a message to a topic
// func (mb *MessageBus) Publish(topic string, msg interface{}) {
// 	if topic == "" || msg == nil {
// 		return
// 	}

// 	if value, exists := mb.topicSubs.Load(topic); exists {
// 		subs := value.([]SubscriptionID)
// 		for _, subID := range subs {
// 			if sub, ok := mb.subscriptions.Load(subID); ok {
// 				subscription := sub.(Subscription)
// 				subscription.Handler(msg)
// 			}
// 		}
// 	}

// 	mb.mu.Lock()
// 	mb.pubCount++
// 	mb.mu.Unlock()
// }

// Send sends a message to a specific endpoint
func (mb *MessageBus) Send(endpoint string, msg interface{}) {
	if endpoint == "" || msg == nil {
		return
	}

	if h, exists := mb.endpoints.Load(endpoint); exists {
		handler := h.(Handler)
		handler(msg)

		mb.mu.Lock()
		mb.sentCount++
		mb.mu.Unlock()
	} else {
		mb.logger.Printf("No endpoint registered at '%s'", endpoint)
	}
}

// Request sends a request and registers its callback
func (mb *MessageBus) Request(endpoint string, req Request) {
	if endpoint == "" || req == nil {
		return
	}

	if _, exists := mb.corrIndex.Load(req.GetID()); exists {
		mb.logger.Printf("Duplicate request ID %s", req.GetID())
		return
	}

	mb.corrIndex.Store(req.GetID(), req.GetCallback())

	if h, exists := mb.endpoints.Load(endpoint); exists {
		handler := h.(Handler)
		handler(req)

		mb.mu.Lock()
		mb.reqCount++
		mb.mu.Unlock()
	} else {
		mb.logger.Printf("No endpoint registered at '%s'", endpoint)
	}
}

// Response handles a response message
func (mb *MessageBus) Response(res Response) {
	if res == nil {
		return
	}

	if h, exists := mb.corrIndex.LoadAndDelete(res.GetCorrelationID()); exists {
		handler := h.(Handler)
		handler(res)

		mb.mu.Lock()
		mb.resCount++
		mb.mu.Unlock()
	} else {
		mb.logger.Printf("No callback found for correlation ID %s", res.GetCorrelationID())
	}
}

// Stats returns message bus statistics
func (mb *MessageBus) Stats() (sent, req, res, pub int64) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return mb.sentCount, mb.reqCount, mb.resCount, mb.pubCount
}
