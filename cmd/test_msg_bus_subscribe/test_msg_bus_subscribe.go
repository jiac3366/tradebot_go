package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"

	"tradebot_go/tradebot/core/messagebus"
	"tradebot_go/tradebot/exchange/binance"
	"tradebot_go/tradebot/logger"
)

var connector *binance.BinancePublicConnector
var err error
var msgBus *messagebus.MessageBus

func init() {
	logger.InitLogger()
	msgBus = messagebus.GetMessageBus("test", uuid.New(), "test", &messagebus.Config{})
	msgBus.Register("bookTicker", msgHandler)
	msgBus.Register("trade", msgHandler)
}

func msgHandler(msg interface{}) {
	fmt.Printf("!!!! msgHandler: %+v\n", msg)
}

func main() {
	connector, err = binance.NewBinancePublicConnector(
		msgBus,
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	if err := connector.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer connector.Close()

	if err := connector.SubscribeBookL1("btcusdt"); err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	if err := connector.SubscribeTrade("btcusdt"); err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Websocket client is running. Press CTRL+C to stop...")
	<-sigChan
	fmt.Println("\nShutting down gracefully...")
	connector.Close()
}
