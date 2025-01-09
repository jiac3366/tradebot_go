package main

import (
	"fmt"

	"path/filepath"
	"runtime"
	"log"

	"tradebot_go/tradebot/base"

	"tradebot_go/tradebot/exchange/binance"
)



func getRootDir() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "../..")
}

func main() {
	// 使用绝对路径读取配置文件
	r := getRootDir()
	configPath := filepath.Join(r, ".keys", "config.yaml")
	fmt.Printf("configPath: %+v\n", configPath)
	config := base.GetConfig(configPath)
	fmt.Printf("config: %+v\n", config)

	client, err := binance.NewClientWithConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	tradeList, err := client.GetTradeList(binance.TradeListParams{
		Symbol: "BTCUSDT",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tradeList)
}
