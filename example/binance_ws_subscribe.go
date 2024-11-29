package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
	"tradebot/tradebot_go/binance"
)

// LatencyCollector 收集延迟统计数据
type LatencyCollector struct {
	stats   map[string]*symbolStats
	statsMu sync.RWMutex
}

type symbolStats struct {
	mu        sync.Mutex
	latencies []float64
}

// NewLatencyCollector 创建新的延迟收集器
func NewLatencyCollector() *LatencyCollector {
	return &LatencyCollector{
		stats: make(map[string]*symbolStats),
	}
}

// WrapHandler 包装原有的消息处理函数，添加延迟统计
func (lc *LatencyCollector) WrapHandler(handler func([]byte) error) func([]byte) error {
	return func(msg []byte) error {
		// 计算延迟
		var trade binance.Trade
		if err := json.Unmarshal(msg, &trade); err != nil {
			return err
		}

		now := time.Now().UnixMilli()
		latency := float64(now - trade.EventTime)

		// 存储延迟数据
		lc.statsMu.Lock()
		if _, exists := lc.stats[trade.Symbol]; !exists {
			lc.stats[trade.Symbol] = &symbolStats{
				latencies: make([]float64, 0, 1000),
			}
		}
		stats := lc.stats[trade.Symbol]
		lc.statsMu.Unlock()

		stats.mu.Lock()
		stats.latencies = append(stats.latencies, latency)
		stats.mu.Unlock()

		// 调用原始处理函数
		return handler(msg)
	}
}

// PrintStats 打印统计信息
func (lc *LatencyCollector) PrintStats() {
	lc.statsMu.RLock()
	defer lc.statsMu.RUnlock()

	for symbol, stats := range lc.stats {
		stats.mu.Lock()
		latencies := make([]float64, len(stats.latencies))
		copy(latencies, stats.latencies)
		stats.mu.Unlock()

		if len(latencies) == 0 {
			continue
		}

		sort.Float64s(latencies)

		// 计算统计数据
		var sum float64
		for _, v := range latencies {
			sum += v
		}
		avg := sum / float64(len(latencies))

		// 计算标准差
		var variance float64
		for _, v := range latencies {
			variance += (v - avg) * (v - avg)
		}
		stdDev := math.Sqrt(variance / float64(len(latencies)))

		fmt.Printf("\nSymbol: %s ", symbol)
		fmt.Printf("Sample size: %d ", len(latencies))
		fmt.Printf("Avg: %.2f ms ", avg)
		fmt.Printf("Median: %.2f ms ", percentile(latencies, 50))
		fmt.Printf("Std Dev: %.2f ms ", stdDev)
		fmt.Printf("95th percentile: %.2f ms ", percentile(latencies, 95))
		fmt.Printf("99th percentile: %.2f ms ", percentile(latencies, 99))
		fmt.Printf("Min: %.2f ms ", latencies[0])
		fmt.Printf("Max: %.2f ms\n", latencies[len(latencies)-1])
		fmt.Println(strings.Repeat("-", 50))
	}
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	rank := p / 100.0 * float64(len(sorted)-1)
	index := int(rank)
	if index >= len(sorted)-1 {
		return sorted[len(sorted)-1]
	}
	fraction := rank - float64(index)
	return sorted[index] + fraction*(sorted[index+1]-sorted[index])
}

// 添加符号列表常量
var symbols = []string{
	"ARKMUSDT",
	"ZECUSDT", "MANTAUSDT", "ENAUSDT", "ARKUSDT",
	"RIFUSDT", "BEAMXUSDT", "METISUSDT", "1000SATSUSDT", "AMBUSDT",
	"CHZUSDT", "RENUSDT", "BANANAUSDT", "TAOUSDT", "RAREUSDT",
	"SXPUSDT", "IDUSDT", "LQTYUSDT", "RPLUSDT", "COMBOUSDT",
	"SEIUSDT", "RDNTUSDT", "BNXUSDT", "NKNUSDT", "BNBUSDT",
	"APTUSDT", "OXTUSDT", "LEVERUSDT", "AIUSDT", "OMNIUSDT",
	"KDAUSDT", "ALPACAUSDT", "STRKUSDT", "FETUSDT", "FIDAUSDT",
	"MKRUSDT", "GMTUSDT", "VIDTUSDT", "UMAUSDT", "RONINUSDT",
	"FIOUSDT", "BALUSDT", "IOUSDT", "LDOUSDT", "KSMUSDT",
	"TURBOUSDT", "GUSDT", "POLUSDT", "XVSUSDT", "SUNUSDT",
	"TIAUSDT", "LRCUSDT", "1MBABYDOGEUSDT", "ZKUSDT", "ZENUSDT",
	"HOTUSDT", "DARUSDT", "AXSUSDT", "TRXUSDT", "LOKAUSDT",
	"LSKUSDT", "GLMUSDT", "ETHFIUSDT", "ONTUSDT", "ONGUSDT",
	"CATIUSDT", "REZUSDT", "NEIROUSDT", "VANRYUSDT", "ANKRUSDT",
	"ALPHAUSDT",
}

func main() {
	collector := NewLatencyCollector()
	client, err := binance.NewWSClient()
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Connect(context.Background()); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 创建交易数据存储
	store, err := binance.NewTradeStore(symbols)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	// Register trade handler
	client.RegisterHandler("trade", collector.WrapHandler(func(msg []byte) error {
		var trade binance.Trade
		if err := json.Unmarshal(msg, &trade); err != nil {
			return err
		}

		// 更新共享内存中的交易数据
		if err := store.UpdateTrade(&trade); err != nil {
			log.Printf("Failed to update trade data: %v", err)
		}

		log.Printf("Trade: Symbol=%s, Price=%s, Quantity=%s",
			trade.Symbol, trade.Price, trade.Quantity)
		return nil
	}))

	// 批量订阅所有交易对
	for _, symbol := range symbols {
		// 转换为小写
		symbolLower := strings.ToLower(symbol)
		if err := client.Subscribe(symbolLower, []string{"trade"}); err != nil {
			log.Printf("Failed to subscribe to %s: %v", symbol, err)
			continue
		}
		log.Printf("Subscribed to %s", symbol)

		// 添加小延迟，避免订阅请求过于频繁 300ms
		time.Sleep(300 * time.Millisecond)
	}

	// sigChan
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 在收到关闭信号后，打印统计信息
	<-sigChan
	log.Println("Shutting down...")
	collector.PrintStats()
}
