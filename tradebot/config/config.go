package config

import (
	"sync"

	log "github.com/BitofferHub/pkg/middlewares/log"
	"github.com/spf13/viper"
)

var config *Config
var configOnce sync.Once

// ExchangeConfig 包含交易所的基础配置
type ExchangeConfig struct {
	APIKey     string `mapstructure:"api_key"`
	SecretKey  string `mapstructure:"secret_key"`
	Passphrase string `mapstructure:"passphrase,omitempty"` // okex 特有的配置
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	Password string `mapstructure:"password"`
}

// Config 总配置结构
type Config struct {
	BinanceFutureTestnet ExchangeConfig `mapstructure:"binance_future_testnet"`
	OkexDemo             ExchangeConfig `mapstructure:"okex_demo"`
	Bybit                ExchangeConfig `mapstructure:"bybit"`
	BybitTestnet2        ExchangeConfig `mapstructure:"bybit_testnet_2"`
	RedisConfig          RedisConfig    `mapstructure:"redis_config"`
}

// LoadConfig loads the configuration using viper
func loadConfig(configPath string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := v.Unmarshal(config); err != nil {
		return nil, err
	}
	return config, nil
}

func GetConfig(configPath string) *Config {
	configOnce.Do(func() {
		c, err := loadConfig(configPath)
		if err != nil {
			log.Error(err)
		}
		config = c
	})
	return config
}
