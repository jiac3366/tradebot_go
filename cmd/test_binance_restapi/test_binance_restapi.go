package main

import (
	"fmt"

	"path/filepath"
	"runtime"
	"log"

	"tradebot_go/tradebot/base"

	"tradebot_go/tradebot/exchange/binance"
	// rest "github.com/adshao/go-binance/v2"
)



func getRootDir() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "../..")
}


func main() {
	// 使用绝对路径读取配置文件
	r := getRootDir()
	configPath := filepath.Join(r, ".keys", "config.yaml")
	config := base.GetConfig(configPath)

	client := binance.NewBinanceClient(config, binance.BinanceAccountTypeUsdMFuturesTestnet)

	tradeList, err := client.GetFApiTradeList(binance.TradeListParams{
		Symbol: "BTCUSDT",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tradeList)
}


