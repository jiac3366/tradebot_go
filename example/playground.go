package main

import (
	"fmt"

	"github.com/google/uuid"
	"tradebot_go/tradebot/binance"
	"tradebot_go/tradebot/core/messagebus"
)

func main() {

	trade := binance.Trade{
		EventType:  "trade",
		EventTime:  1709616000000, // 2024-03-05 12:00:00 UTC
		Symbol:     "BTCUSDT",
		TradeID:    12345678,
		Price:      "67241.50",
		Quantity:   "0.15623",
		TradeTime:  1709616000000,
		IsMaker:    false,
		Ignore:     false,
		MarketType: "SPOT",
	}

	// 创建一个 BookTicker 示例
	bookTicker := binance.BookTicker{
		UpdateID: 400900217,
		Symbol:   "BTCUSDT",
		BidPrice: "67240.50",
		BidQty:   "1.25000",
		AskPrice: "67241.50",
		AskQty:   "0.84320",
	}

	mb := messagebus.NewMessageBus(
		"TRADER-001",
		uuid.New(),
		"MyMessageBus",
		&messagebus.Config{
			BufferIntervalMS: 100,
		},
	)

	// 注册端点处理器
	mb.Register("bookTicker", func(msg interface{}) {
		bookTicker := msg.(binance.BookTicker)
		fmt.Printf("chase Received bookTicker: %v\n", bookTicker)
	})

	// 订阅主题
	mb.Register("trades.*", func(msg interface{}) {
		trade := msg.(binance.Trade)
		fmt.Printf("chase Received trade: %v\n", trade)
	})

	// 发布消息
	mb.Send("trades.*", trade)

	// 发送消息到端点
	mb.Send("bookTicker", bookTicker)
}
