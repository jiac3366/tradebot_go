package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tradebot/tradebot_go/binance"
)

func main() {
	ctx := context.Background()
	client, err := binance.NewWSClient()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to WebSocket
	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Register trade handler
	client.RegisterHandler("trade", func(msg []byte) error {
		var trade binance.Trade
		if err := json.Unmarshal(msg, &trade); err != nil {
			return err
		}
		log.Printf("Trade: Symbol=%s, Price=%s, Quantity=%s",
			trade.Symbol, trade.Price, trade.Quantity)
		return nil
	})

	// Subscribe to BTCUSDT trades
	if err := client.Subscribe("btcusdt", []string{"trade"}); err != nil {
		log.Fatal(err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down...")

	// Give some time for cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Close(); err != nil {
		log.Printf("Error closing connection: %v", err)
	}
}
