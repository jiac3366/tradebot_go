package binance

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestBinanceWSClientConnect(t *testing.T) {
	// Create message handler
	handler := func(msg map[string]interface{}) error {
		fmt.Printf("%+v\n", msg)
		return nil
	}

	client, err := NewBinanceWSClient(BinanceAccountTypeUsdMFuturesTestnet, handler)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	client.Subscribe("btcusdt", "trade")

	time.Sleep(15 * time.Second)

	if err := client.Close(); err != nil {
		t.Errorf("failed to close connection: %v", err)
	}
}
