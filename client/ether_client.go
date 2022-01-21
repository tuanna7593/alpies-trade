package client

import (
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	clientOnce         sync.Once
	ethClientSingleton *ethclient.Client
)

// SetEtheClient init the ethereum client
func SetEtheClient(c *ethclient.Client) *ethclient.Client {
	clientOnce.Do(func() {
		ethClientSingleton = c
	})
	return ethClientSingleton
}

// GetEtheClient gets the instance of ethereum client
func GetEtheClient() *ethclient.Client {
	return ethClientSingleton
}
