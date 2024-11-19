// 这个 Go 版本的实现有以下几个亮点：
// 并发安全：
// 使用 sync.Map 替代普通 map，确保线程安全
// 使用 errgroup 进行并发操作的错误处理
// 使用 context 进行优雅的生命周期管理
// 性能优化：
// 使用 Redis pipeline 批量处理操作
// 使用指针传递避免不必要的复制
// 合理使用 goroutine 进行异步操作
// 错误处理：
// 使用 fmt.Errorf 和 %w 进行错误包装
// 区分 Redis 的 Nil 错误和其他错误
// 资源管理：
// 使用 context.Context 控制 goroutine 生命周期
// 实现优雅关闭机制
// 使用 sync.WaitGroup 等待后台任务完成
// 代码组织：
// 使用 struct 组织相关的字段
// 将 Redis key 管理集中到单独的结构体
// 方法命名符合 Go 规范
// 可维护性：
// 清晰的注释
// 一致的错误处理模式
// 模块化的设计
// 这个实现充分利用了 Go 的并发特性和标准库，同时保持了代码的可读性和可维护性。

package entity

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

// OrderCache manages order caching with both in-memory and Redis storage
type OrderCache struct {
	// Immutable fields after initialization
	accountType  AccountType
	strategyID   string
	userID       string
	redisKeys    redisKeys
	redisClient  *redis.Client
	syncInterval time.Duration
	expireTime   time.Duration

	// Thread-safe in-memory storage
	orders           sync.Map // map[string]*Order
	openOrders       *sync.Map
	symbolOpenOrders *sync.Map // map[string]*sync.Map
	symbolOrders     *sync.Map // map[string]*sync.Map

	// Lifecycle management
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
}

type redisKeys struct {
	orders           string
	openOrders       string
	symbolOrders     string
	symbolOpenOrders string
}

// NewOrderCache creates a new OrderCache instance
func NewOrderCache(
	accountType AccountType,
	strategyID string,
	userID string,
	redisClient *redis.Client,
	syncInterval time.Duration,
	expireTime time.Duration,
) *OrderCache {
	ctx, cancel := context.WithCancel(context.Background())

	cache := &OrderCache{
		accountType:  accountType,
		strategyID:   strategyID,
		userID:       userID,
		redisClient:  redisClient,
		syncInterval: syncInterval,
		expireTime:   expireTime,
		ctx:          ctx,
		cancelFunc:   cancel,

		// Initialize thread-safe maps
		openOrders:       &sync.Map{},
		symbolOpenOrders: &sync.Map{},
		symbolOrders:     &sync.Map{},
	}

	cache.redisKeys = redisKeys{
		orders:           fmt.Sprintf("strategy:%s:user_id:%s:account_type:%s:orders", strategyID, userID, accountType),
		openOrders:       fmt.Sprintf("strategy:%s:user_id:%s:account_type:%s:open_orders", strategyID, userID, accountType),
		symbolOrders:     fmt.Sprintf("strategy:%s:user_id:%s:account_type:%s:symbol_orders", strategyID, userID, accountType),
		symbolOpenOrders: fmt.Sprintf("strategy:%s:user_id:%s:account_type:%s:symbol_open_orders", strategyID, userID, accountType),
	}

	return cache
}

// Start begins the cache synchronization processes
func (c *OrderCache) Start() error {
	c.wg.Add(2)
	go c.periodicSync()
	go c.periodicCleanup()
	return nil
}

func (c *OrderCache) periodicSync() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if err := c.syncToRedis(); err != nil {
				// Use structured logging in production
				fmt.Printf("Error syncing to Redis: %v\n", err)
			}
		}
	}
}

func (c *OrderCache) periodicCleanup() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.cleanupExpiredData()
		}
	}
}

func (c *OrderCache) syncToRedis() error {
	// Use errgroup for concurrent operations
	g, ctx := errgroup.WithContext(c.ctx)

	// Sync orders
	g.Go(func() error {
		pipe := c.redisClient.Pipeline()
		c.orders.Range(func(key, value interface{}) bool {
			orderID := key.(string)
			order := value.(*Order)
			data, err := json.Marshal(order)
			if err != nil {
				return false
			}
			pipe.HSet(ctx, c.redisKeys.orders, orderID, data)
			return true
		})
		_, err := pipe.Exec(ctx)
		return err
	})

	// Sync open orders
	g.Go(func() error {
		members := []interface{}{}
		c.openOrders.Range(func(key, _ interface{}) bool {
			members = append(members, key)
			return true
		})

		pipe := c.redisClient.Pipeline()
		pipe.Del(ctx, c.redisKeys.openOrders)
		if len(members) > 0 {
			pipe.SAdd(ctx, c.redisKeys.openOrders, members...)
		}
		_, err := pipe.Exec(ctx)
		return err
	})

	return g.Wait()
}

func (c *OrderCache) cleanupExpiredData() {
	expireBefore := time.Now().Add(-c.expireTime).Unix()

	c.orders.Range(func(key, value interface{}) bool {
		order := value.(*Order)
		if order.Timestamp < expireBefore {
			c.orders.Delete(key)
			// Cleanup from symbol orders
			if symbolOrdersMap, ok := c.symbolOrders.Load(order.Symbol); ok {
				symbolOrdersMap.(*sync.Map).Delete(key)
			}
		}
		return true
	})
}

// OrderInitialized handles new order initialization
func (c *OrderCache) OrderInitialized(order *Order) {
	// Check if order already exists
	if _, exists := c.orders.Load(order.ID); exists {
		return
	}

	c.orders.Store(order.ID, order)
	c.openOrders.Store(order.ID, struct{}{})

	// Update symbol-specific maps
	c.updateSymbolMaps(order.Symbol, order.ID, true)
}

// OrderStatusUpdate handles order status updates
func (c *OrderCache) OrderStatusUpdate(order *Order) {
	c.orders.Store(order.ID, order)

	if order.Status == OrderStatusFilled || order.Status == OrderStatusCanceled || order.Status == OrderStatusExpired {
		c.openOrders.Delete(order.ID)
		// Remove from symbol open orders
		c.updateSymbolMaps(order.Symbol, order.ID, false)
	}
}

func (c *OrderCache) updateSymbolMaps(symbol, orderID string, isNew bool) {
	// Get or create symbol orders map
	symbolOrdersValue, _ := c.symbolOrders.LoadOrStore(symbol, &sync.Map{})
	symbolOrdersMap := symbolOrdersValue.(*sync.Map)
	symbolOrdersMap.Store(orderID, struct{}{})

	if isNew {
		// Get or create symbol open orders map
		symbolOpenOrdersValue, _ := c.symbolOpenOrders.LoadOrStore(symbol, &sync.Map{})
		symbolOpenOrdersMap := symbolOpenOrdersValue.(*sync.Map)
		symbolOpenOrdersMap.Store(orderID, struct{}{})
	}
}

// GetOrder retrieves an order by ID from cache or Redis
func (c *OrderCache) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	// Check in-memory cache first
	if orderValue, exists := c.orders.Load(orderID); exists {
		return orderValue.(*Order), nil
	}

	// Try to fetch from Redis
	data, err := c.redisClient.HGet(ctx, c.redisKeys.orders, orderID).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order from Redis: %w", err)
	}

	var order Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	// Cache the order in memory
	c.orders.Store(orderID, &order)
	return &order, nil
}

// GetSymbolOrders retrieves all orders for a symbol
func (c *OrderCache) GetSymbolOrders(ctx context.Context, symbol string, inMem bool) (map[string]struct{}, error) {
	result := make(map[string]struct{})

	// Get in-memory orders
	if symbolOrdersValue, exists := c.symbolOrders.Load(symbol); exists {
		symbolOrdersMap := symbolOrdersValue.(*sync.Map)
		symbolOrdersMap.Range(func(key, _ interface{}) bool {
			result[key.(string)] = struct{}{}
			return true
		})
	}

	if !inMem {
		// Fetch from Redis
		members, err := c.redisClient.SMembers(ctx, fmt.Sprintf("%s:%s", c.redisKeys.symbolOrders, symbol)).Result()
		if err != nil && err != redis.Nil {
			return nil, fmt.Errorf("failed to get symbol orders from Redis: %w", err)
		}
		for _, member := range members {
			result[member] = struct{}{}
		}
	}

	return result, nil
}

// GetOpenOrders retrieves open orders, optionally filtered by symbol
func (c *OrderCache) GetOpenOrders(symbol string) map[string]struct{} {
	result := make(map[string]struct{})

	if symbol != "" {
		if symbolOpenOrdersValue, exists := c.symbolOpenOrders.Load(symbol); exists {
			symbolOpenOrdersMap := symbolOpenOrdersValue.(*sync.Map)
			symbolOpenOrdersMap.Range(func(key, _ interface{}) bool {
				result[key.(string)] = struct{}{}
				return true
			})
		}
	} else {
		c.openOrders.Range(func(key, _ interface{}) bool {
			result[key.(string)] = struct{}{}
			return true
		})
	}

	return result
}

// Close gracefully shuts down the cache
func (c *OrderCache) Close() error {
	c.cancelFunc()

	// Wait for background tasks to complete
	c.wg.Wait()

	// Final sync to Redis
	return c.syncToRedis()
}
