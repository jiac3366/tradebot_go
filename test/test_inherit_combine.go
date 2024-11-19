package main

import (
	"fmt"
	"time"
)

// 基础结构体
type BaseEntity struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BaseEntity 的方法
func (b *BaseEntity) GetID() string {
	return b.ID
}

func (b *BaseEntity) SetUpdatedAt() {
	b.UpdatedAt = time.Now()
	fmt.Println("BaseEntity.SetUpdatedAt called")
}

// 价格接口
type Pricer interface {
	GetPrice() float64
	SetPrice(float64)
}

// 交易结构体（通过组合获得 BaseEntity 的字段和方法）
type Trade struct {
	BaseEntity         // 嵌入 BaseEntity
	Symbol     string  // 交易对
	Price      float64 // 价格
	Quantity   float64 // 数量
	Direction  string  // 买/卖方向
}

// Trade 特有的方法
func (t *Trade) GetPrice() float64 {
	return t.Price
}

func (t *Trade) SetPrice(price float64) {
	t.Price = price
	t.SetUpdatedAt() // 可以直接调用 BaseEntity 的方法
}

func (t *Trade) GetValue() float64 {
	return t.Price * t.Quantity
}

func (t *Trade) SetUpdatedAt() {
	t.BaseEntity.SetUpdatedAt() // 可以显式调用 BaseEntity 的方法
	fmt.Println("Trade.SetUpdatedAt called")
}

// 订单结构体（同样组合 BaseEntity）
type Order struct {
	entity   BaseEntity // 嵌入 BaseEntity
	Symbol   string
	Price    float64
	Quantity float64
	Status   string // 订单状态
}

// Order 特有的方法
func (o *Order) GetPrice() float64 {
	return o.Price
}

func (o *Order) SetPrice(price float64) {
	o.Price = price
	o.entity.SetUpdatedAt()
}

func (o *Order) Cancel() {
	o.Status = "cancelled"
	o.entity.SetUpdatedAt()
}

// 交易服务
type TradeService struct {
	logger Logger // 组合 Logger 接口
}

// Logger 接口
type Logger interface {
	Log(message string)
}

// 具体的日志实现
type ConsoleLogger struct{}

func (l ConsoleLogger) Log(message string) {
	fmt.Println("LOG:", message)
}

// 创建新的交易服务
func NewTradeService(logger Logger) *TradeService {
	return &TradeService{
		logger: logger,
	}
}

// 处理价格更新
func (s *TradeService) ProcessPriceUpdate(p Pricer, newPrice float64) {
	oldPrice := p.GetPrice()
	p.SetPrice(newPrice)
	s.logger.Log(fmt.Sprintf("Price updated from %.2f to %.2f", oldPrice, newPrice))
}

func main() {
	// 创建服务
	service := NewTradeService(ConsoleLogger{})

	// 创建交易
	trade := &Trade{
		BaseEntity: BaseEntity{
			ID:        "T001",
			CreatedAt: time.Now(),
		},
		Symbol:    "BTCUSDT",
		Price:     30000,
		Quantity:  1.5,
		Direction: "buy",
	}

	// 创建订单
	order := &Order{
		entity: BaseEntity{
			ID:        "O001",
			CreatedAt: time.Now(),
		},
		Symbol:   "BTCUSDT",
		Price:    30000,
		Quantity: 1.5,
		Status:   "new",
	}

	trade.SetUpdatedAt()

	// 使用接口处理不同类型
	service.ProcessPriceUpdate(trade, 31000)
	service.ProcessPriceUpdate(order, 31000)

	// 访问嵌入字段的方法
	fmt.Println("Trade ID:", trade.GetID())
	fmt.Println("Trade Value:", trade.GetValue())
	fmt.Println("Order ID:", order.entity.GetID())

	// 取消订单
	order.Cancel()
	fmt.Println("Order Status:", order.Status)
}
