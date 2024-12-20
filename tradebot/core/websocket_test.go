package core

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// setupTestServer creates a test WebSocket server
func setupTestServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, string) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("failed to upgrade connection: %v", err)
			return
		}
		defer conn.Close()

		handler(w, r)
	}))

	// Convert http://... to ws://...
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	return server, wsURL
}

func TestNewWSClient(t *testing.T) {
	handler := func(msg map[string]interface{}) error { return nil }
	client, err := NewWSClient("ws://test.com", handler)

	if err != nil {
		t.Errorf("NewWSClient failed: %v", err)
	}
	if client == nil {
		t.Error("NewWSClient returned nil client")
	}
	if client.status != "disconnected" {
		t.Errorf("expected status 'disconnected', got '%s'", client.status)
	}
}

func TestWSClientPing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		// Set up ping handler
		conn.SetPingHandler(func(string) error {
			fmt.Println("ping received")
			return conn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(time.Second))
		})

		// Keep connection alive
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer server.Close()

	// Convert http to ws
	url := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create client with handler
	client, err := NewWSClient(url, func(msg map[string]interface{}) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	client.Connect(context.Background())
	// Test ping
	go client.Ping(3 * time.Second)

	time.Sleep(10 * time.Second)
}

// func TestWSClientReconnect(t *testing.T) {
// 	server, wsURL := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
// 		conn, _ := upgrader.Upgrade(w, r, nil)
// 		conn.WriteJSON(map[string]string{"type": "test"})
// 	})
// 	defer server.Close()

// 	client, _ := NewWSClient(wsURL, func(msg map[string]interface{}) error { return nil })

// 	// Test initial connection
// 	err := client.Connect(context.Background())
// 	if err != nil {
// 		t.Fatalf("initial connect failed: %v", err)
// 	}

// 	// Test close and reconnect
// 	client.Close()
// 	err = client.reconnect()
// 	if err != nil {
// 		t.Errorf("reconnect failed: %v", err)
// 	}
// }
