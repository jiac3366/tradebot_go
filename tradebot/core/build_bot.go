package core

import (
	"path/filepath"
	"runtime"
	"sync"

	log "github.com/BitofferHub/pkg/middlewares/log"
)

var (
	rootPath     string
	rootPathOnce sync.Once
)

// GetRootPath 获取项目根目录
func GetRootPath() string {
	rootPathOnce.Do(func() {
		// 获取当前文件的路径
		_, currentFile, _, _ := runtime.Caller(0)
		// 获取 core 目录的父目录，即项目根目录
		rootPath = filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))
	})
	return rootPath
}

func InitLogger() {
	// 使用 GetRootPath() 构建日志路径
	logPath := filepath.Join(GetRootPath(), ".log")

	log.Init(
		log.WithLogPath(logPath),
		log.WithLogLevel("info"),
		log.WithFileName("tradebot-go.log"),
		log.WithMaxBackups(100),
		log.WithMaxSize(1024*1024*10),
		log.WithConsole(false),
	)
}

func init() {
	InitLogger()
}
