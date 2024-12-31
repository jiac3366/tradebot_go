package main

import (
	"context"
	"fmt"
	"time"

	"tradebot_go/tradebot/binance"
)

func HandleTradeMessage(msg map[string]interface{}) (*binance.Trade, error) {
	// 将map数据转换为Trade结构体
	trade := &binance.Trade{}

	// 使用类型断言获取数据
	if e, ok := msg["e"].(string); ok {
		trade.EventType = e
	}
	if t, ok := msg["E"].(float64); ok {
		trade.EventTime = int64(t)
	}
	if s, ok := msg["s"].(string); ok {
		trade.Symbol = s
	}
	if tid, ok := msg["t"].(float64); ok {
		trade.TradeID = int64(tid)
	}
	if p, ok := msg["p"].(string); ok {
		trade.Price = p
	}
	if q, ok := msg["q"].(string); ok {
		trade.Quantity = q
	}

	if tt, ok := msg["T"].(float64); ok {
		trade.TradeTime = int64(tt)
	}
	if m, ok := msg["m"].(bool); ok {
		trade.IsMaker = m
	}
	if x, ok := msg["X"].(bool); ok {
		trade.MarketType = x
	}
	return trade, nil
}

func main() {
	client, err := binance.NewBinanceWSClient(binance.BinanceAccountTypeUsdMFuturesTestnet, func(msg map[string]interface{}) error {
		trade, err := HandleTradeMessage(msg)
		if err != nil {
			return err
		}
		// 计算延迟
		delay := time.Now().UnixMilli() - trade.EventTime
		fmt.Printf("Trade: Symbol=%s, Price=%s, Quantity=%s, IsBuyerMaker=%v, Delay=%dms\n",
			trade.Symbol, trade.Price, trade.Quantity, trade.IsMaker, delay)
		return nil
	})
	if err != nil {
		fmt.Printf("failed to create client: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		fmt.Printf("failed to connect: %v", err)
		return
	}

	client.Subscribe("btcusdt", "trade")
	time.Sleep(15 * time.Second)
}
