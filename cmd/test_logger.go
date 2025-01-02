package main

import (
	
	log "github.com/BitofferHub/pkg/middlewares/log"
)

func InitLogger() {
	log.Init(
		log.WithLogPath("./log/"),
		log.WithLogLevel("info"),
		log.WithFileName("tradebot-go.log"),
		log.WithMaxBackups(100),
		log.WithMaxSize(1024*1024*10),
		log.WithConsole(false),
	)
}

func main() {
	InitLogger()
	log.Infof("test")
	log.Infof("test")
}
