package main

import (
    "fmt"
    "runtime"
    "time"
)

// 模拟一个大型的基础客户端
type BaseClient struct {
    ApiKey    string
    SecretKey string
    Data      map[string]interface{}
    BigSlice  []int    // 用于占用更多内存
    Time      []time.Time
}

// 使用值复制的客户端
type ValueClient struct {
    BaseClient
    ExID       string
}

// 使用指针的客户端
type PointerClient struct {
    *BaseClient
    ExID       string
}

// 打印当前内存使用情况
func printMemUsage() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
    fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
    fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
    fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
    return b / 1024 / 1024
}

func main() {
    // 创建一个较大的基础客户端
    baseClient := BaseClient{
        ApiKey:    "very_long_api_key_" + string(make([]byte, 1000)),
        SecretKey: "very_long_secret_key_" + string(make([]byte, 1000)),
        Data:      make(map[string]interface{}),
        BigSlice:  make([]int, 10000000),     // 分配大量内存
        Time:      make([]time.Time, 1000000), // 更多内存分配
    }

    // 填充一些数据
    for i := 0; i < 10000000; i++ {
        baseClient.BigSlice[i] = i
    }
    for i := 0; i < 1000000; i++ {
        baseClient.Time[i] = time.Now()
    }

    fmt.Println("Initial memory usage:")
    printMemUsage()

    // 测试值复制
    fmt.Println("\nCreating  clients with value copying:")
    valueClients := make([]ValueClient, 1000000)
    for i := 0; i < 1000000; i++ {
        valueClients[i] = ValueClient{
            BaseClient: baseClient,
            ExID:      fmt.Sprintf("client_%d", i),
        }
    }
    printMemUsage()

    // 清理以准备下一个测试
    valueClients = nil
    runtime.GC()
    time.Sleep(time.Second) // 等待GC完成

    fmt.Println("\nAfter clearing value clients:")
    printMemUsage()

    // 测试指针
    fmt.Println("\nCreating  clients with pointers:")
    pointerClients := make([]PointerClient, 1000000)
    for i := 0; i < 1000000; i++ {
        pointerClients[i] = PointerClient{
            BaseClient: &baseClient,
            ExID:      fmt.Sprintf("client_%d", i),
        }
    }
    printMemUsage()
}