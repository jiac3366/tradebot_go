package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tradebot_go/tradebot/binance"
)

// 添加一个全局变量来存储 client
var wsClient *binance.BinanceWSClient

func handleTradeStream(msg map[string]interface{}) error {
	fmt.Printf("Received raw message: %+v\n", msg)
	trade, err := wsClient.HandleTradeMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to handle trade message: %w", err)
	}

	// 计算延迟
	delay := time.Now().UnixMilli() - trade.EventTime
	fmt.Printf("Trade: Symbol=%s, Price=%s, Quantity=%s, IsBuyerMaker=%v, Delay=%dms\n",
		trade.Symbol, trade.Price, trade.Quantity, trade.IsMaker, delay)
	return nil
}

func main() {
	var err error
	// 将 client 赋值给全局变量
	wsClient, err = binance.NewBinanceWSClient(
		binance.BinanceAccountTypeUsdMFuturesTestnet,
		handleTradeStream,
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// // 设置连接超时
	// ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	// defer cancel()

	// 建立连接
	if err := wsClient.Connect(context.Background()); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer wsClient.Close()

	// 订阅交易流
	if err := wsClient.Subscribe("btcusdt", "trade"); err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	time.Sleep(15 * time.Second)
	wsClient.Close()
	// 创建一个通道来处理信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 添加必要的导入
	fmt.Println("Websocket client is running. Press CTRL+C to stop...")

	// 等待信号而不是使用 sleep
	<-sigChan

	fmt.Println("\nShutting down gracefully...")
	wsClient.Close()
}
