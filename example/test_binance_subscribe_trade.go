package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"tradebot_go/tradebot/binance"
)

func handleTradeStream(msg map[string]interface{}) error {
	// 使用 BinanceWSClient 的 HandleTradeMessage 方法
	client := &binance.BinanceWSClient{}
	trade, err := client.HandleTradeMessage(msg)
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
	// 创建 WebSocket 客户端
	client, err := binance.NewBinanceWSClient(
		binance.BinanceAccountTypeUsdMFuturesTestnet,
		handleTradeStream,
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// 设置连接超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 建立连接
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// 订阅交易流
	if err := client.Subscribe("btcusdt", "trade"); err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	// 运行 15 秒后退出
	time.Sleep(30 * time.Second)
}
