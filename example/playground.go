package main

import (
	"context"
	"fmt"
	"golang.org/x/time/rate"
	"time"
)

func main() {
	// 创建一个限流器：每300ms允许1个token
	limiter := rate.NewLimiter(rate.Every(300*time.Millisecond), 1)
	
	// 运行10次来测试
	for i := 0; i < 10; i++ {
		// Wait会阻塞直到获得一个token
		limiter.Wait(context.Background())
		fmt.Printf("打印第 %d 次: %s\n", i+1, time.Now().Format("15:04:05.000"))
	}
}