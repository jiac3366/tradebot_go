package binance

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"syscall"
)

const (
	// 每个交易数据的大小
	TradeDataSize = 256 // 预留足够空间给每个交易对的数据
	// 共享内存文件路径
	ShmPath = "/dev/shm/binance_trades"
)

// TradeData 表示共享内存中的交易数据结构
type TradeData struct {
	Symbol    [32]byte  // 固定长度的符号名
	Price     [32]byte  // 价格字符串
	Quantity  [32]byte  // 数量字符串
	EventTime int64     // 事件时间戳
	TradeTime int64     // 交易时间戳
	TradeID   int64     // 交易ID
	IsMaker   byte      // 是否是maker
	Padding   [119]byte // 填充到256字节
}

// TradeStore 管理共享内存中的交易数据
type TradeStore struct {
	mu       sync.RWMutex
	data     []byte // mmap的内存
	file     *os.File
	symbols  map[string]int // 符号到偏移量的映射
	numTrade int            // 交易对数量
}

// NewTradeStore 创建新的交易数据存储
func NewTradeStore(symbols []string) (*TradeStore, error) {
	numTrade := len(symbols)
	size := numTrade * TradeDataSize

	// 创建或打开共享内存文件
	file, err := os.OpenFile(ShmPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// 设置文件大小
	if err := file.Truncate(int64(size)); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to truncate: %w", err)
	}

	// 映射到内存
	data, err := syscall.Mmap(int(file.Fd()), 0, size,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to mmap: %w", err)
	}

	// 创建符号映射
	symbolMap := make(map[string]int, numTrade)
	for i, symbol := range symbols {
		symbolMap[symbol] = i * TradeDataSize
	}

	return &TradeStore{
		data:     data,
		file:     file,
		symbols:  symbolMap, // dict which stores the index of the symbol
		numTrade: numTrade,  // length of the symbols
	}, nil
}

// UpdateTrade 更新交易数据
func (ts *TradeStore) UpdateTrade(trade *Trade) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	offset, ok := ts.symbols[trade.Symbol]
	if !ok {
		return fmt.Errorf("unknown symbol: %s", trade.Symbol)
	}

	// 将交易数据写入共享内存
	var data TradeData
	copy(data.Symbol[:], trade.Symbol)
	copy(data.Price[:], trade.Price)
	copy(data.Quantity[:], trade.Quantity)
	data.EventTime = trade.EventTime
	data.TradeTime = trade.TradeTime
	data.TradeID = trade.TradeID
	if trade.IsMaker {
		data.IsMaker = 1
	}

	// 写入内存
	return ts.writeTradeData(offset, &data)
}

// writeTradeData 将交易数据写入指定偏移量
func (ts *TradeStore) writeTradeData(offset int, data *TradeData) error {
	// 写入符号
	copy(ts.data[offset:], data.Symbol[:])
	offset += 32

	// 写入价格
	copy(ts.data[offset:], data.Price[:])
	offset += 32

	// 写入数量
	copy(ts.data[offset:], data.Quantity[:])
	offset += 32

	// 写入时间戳和其他数据
	binary.LittleEndian.PutUint64(ts.data[offset:], uint64(data.EventTime))
	offset += 8
	binary.LittleEndian.PutUint64(ts.data[offset:], uint64(data.TradeTime))
	offset += 8
	binary.LittleEndian.PutUint64(ts.data[offset:], uint64(data.TradeID))
	offset += 8
	ts.data[offset] = data.IsMaker

	return nil
}

// Close 关闭存储
func (ts *TradeStore) Close() error {
	if err := syscall.Munmap(ts.data); err != nil {
		return fmt.Errorf("failed to unmap: %w", err)
	}
	return ts.file.Close()
}
