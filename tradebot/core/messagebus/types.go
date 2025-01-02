package messagebus

import (
	"github.com/google/uuid"
)

// Handler represents a message handler function
type Handler func(msg interface{})

// Request represents a request message with callback
type Request interface {
	GetID() uuid.UUID
	GetCallback() Handler
}

// Response represents a response message
type Response interface {
	GetCorrelationID() uuid.UUID
}

// Subscription represents a subscription to a topic
type Subscription struct {
	ID       SubscriptionID
	Topic    string
	Handler  Handler
	Priority int
}
// Config represents message bus configuration
type Config struct {
	// Database configuration
	Database string
	// Message encoding format
	Encoding string
	// Use ISO8601 timestamps
	TimestampsAsISO8601 bool
	// Buffer interval in milliseconds
	BufferIntervalMS int
	// Auto trim interval in minutes
	AutoTrimMins int
	// Use trader prefix
	UseTraderPrefix bool
	// Use trader ID
	UseTraderID bool
	// Use instance ID
	UseInstanceID bool
	// Streams prefix
	StreamsPrefix string
	// Types filter
	TypesFilter []string
}
