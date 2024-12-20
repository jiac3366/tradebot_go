package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"tradebot_go/tradebot/core/logger"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger 结合两者优点
func InitLogger(logDir string) error {
	// 1. 使用 lumberjack 处理日志轮转
	rotator := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "tradebot.log"),
		MaxSize:    100,  // 每个文件最大100MB
		MaxBackups: 7,    // 保留7个备份
		MaxAge:     7,    // 保留7天
		Compress:   true, // 压缩旧文件
		LocalTime:  true,
	}

	// 2. 使用 logrus 的高级特性
	logger.Logger.SetOutput(rotator)
	logger.Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := filepath.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})

	// 3. 添加自定义 Hook
	logger.Logger.AddHook(&contextHook{})

	return nil
}

// contextHook 为每条日志添加额外信息
type contextHook struct{}

func (hook *contextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *contextHook) Fire(entry *logrus.Entry) error {
	if entry.Data == nil {
		entry.Data = make(logrus.Fields)
	}
	entry.Data["hostname"], _ = os.Hostname()
	entry.Data["pid"] = os.Getpid()
	return nil
}

// 使用示例
func main() {
	// 基础日志
	logger.Logger.Info("Simple log message")

	// 带字段的结构化日志
	logger.Logger.WithFields(logrus.Fields{
		"symbol": "BTCUSDT",
		"price":  30000,
		"action": "trade",
	}).Info("Trade executed")

	// 错误日志
	err := fmt.Errorf("connection failed")
	logger.Logger.WithError(err).Error("Database error")

	// 性能日志
	logger.Logger.WithFields(logrus.Fields{
		"latency_ms": 100,
		"endpoint":   "/api/v1/trades",
	}).Debug("API request completed")
}
