package es

import (
	"fmt"
	"sync"
)

var lock = &sync.Mutex{}

var DataProviders = map[string]ConnFactory{}
var StreamProviders = map[string]StreamerFactory{}

func RegisterDataProviders(name string, factory ConnFactory) {
	lock.Lock()
	defer lock.Unlock()

	DataProviders[name] = factory
}

func RegisterStreamProviders(name string, factory StreamerFactory) {
	lock.Lock()
	defer lock.Unlock()

	StreamProviders[name] = factory
}

func GetConn(cfg *ProviderConfig) (Conn, error) {
	lock.Lock()
	defer lock.Unlock()

	if factory, ok := DataProviders[cfg.Data.Type]; ok {
		return factory(cfg.Data)
	}

	return nil, fmt.Errorf("data provider not found: %s", cfg.Data.Type)
}

func GetStreamer(cfg *ProviderConfig) (Streamer, error) {
	lock.Lock()
	defer lock.Unlock()

	if factory, ok := StreamProviders[cfg.Stream.Type]; ok {
		return factory(cfg.Stream)
	}

	return nil, fmt.Errorf("streamer provider not found: %s", cfg.Stream.Type)
}
