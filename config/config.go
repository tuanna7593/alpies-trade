package config

import "sync"

type ClientConfig struct {
	RPCEndpoint      string `mapstructure:"RPC_ENDPOINT"`
	NFTMarketAddress string `mapstructure:"NFT_MARKET_ADDRESS"`
	AlpiesAddress    string `mapstructure:"ALPIES_ADDRESS"`
}

var (
	cfgOnce      sync.Once
	cfgSingleton *ClientConfig
)

// SetConfig to set configuration of service.
func InitConfig(cfg *ClientConfig) *ClientConfig {
	cfgOnce.Do(func() {
		cfgSingleton = cfg
	})
	return cfgSingleton
}

// GetConfig gets the instance of singleton
func GetConfig() *ClientConfig {
	return cfgSingleton
}
