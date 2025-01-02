package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tradebot_go/tradebot/exchange/binance"
)

var wsClient *binance.BinanceWSClient

func handleBookL1Stream(msg map[string]interface{}) error {
	bookL1, err := wsClient.HandleBookL1Message(msg)
	if err != nil {
		return fmt.Errorf("failed to handle trade message: %w", err)
	}

	fmt.Printf("BookL1: Symbol=%s, BidPrice=%s, BidQty=%s, AskPrice=%s, AskQty=%s\n",
		bookL1.Symbol, bookL1.BidPrice, bookL1.BidQty, bookL1.AskPrice, bookL1.AskQty)
	return nil
}

func handleTradeStream(msg map[string]interface{}) error {
	trade, err := wsClient.HandleTradeMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to handle trade message: %w", err)
	}

	// 计算延迟
	delay := time.Now().UnixMilli() - trade.EventTime
	fmt.Printf("Trade: Symbol=%s, Price=%s, Quantity=%s, IsMaker=%v, Delay=%dms\n",
		trade.Symbol, trade.Price, trade.Quantity, trade.IsMaker, delay)
	return nil
}


func main() {
	var err error
	// 将 client 赋值给全局变量
	wsClient, err = binance.NewBinanceWSClient(
		binance.BinanceAccountTypeUsdMFuturesTestnet,
		handleBookL1Stream,
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// 建立连接
	if err := wsClient.Connect(context.Background()); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer wsClient.Close()

	// 订阅交易流
	if err := wsClient.SubscribeBookL1("btcusdt"); err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	

	time.Sleep(15 * time.Second)
	// test
	wsClient.CloseConnection()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Websocket client is running. Press CTRL+C to stop...")
	<-sigChan
	fmt.Println("\nShutting down gracefully...")
	wsClient.Close()
}
