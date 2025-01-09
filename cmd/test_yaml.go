package main

import (
	"fmt"
	"log"

	// "tradebot_go/tradebot/exchange/binance"
	"tradebot_go/tradebot/config"
	"encoding/json"
)

// func main() {
// 	// 从配置文件创建客户端
// 	client, err := binance.NewClientWithConfig(".keys/config.yaml")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	params := binance.TradeListParams{
// 		Symbol: "BTCUSDT",
// 	}

// 	trades, err := client.GetTradeList(params)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	for _, trade := range trades {
// 		fmt.Printf("Trade ID: %d, Price: %s\n", trade.ID, trade.Price)
// 	}
// }

func main() {
	config, err := config.LoadConfig(".keys/config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", config)

	// to json
	jsonConfig, err := json.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(jsonConfig))
}
