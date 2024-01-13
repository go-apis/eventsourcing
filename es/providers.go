package es

import (
	"context"
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

func GetConn(ctx context.Context, cfg *ProviderConfig, reg Registry) (Conn, error) {
	lock.Lock()
	defer lock.Unlock()

	if factory, ok := DataProviders[cfg.Data.Type]; ok {
		return factory(ctx, cfg, reg)
	}

	return nil, fmt.Errorf("data provider not found: %s", cfg.Data.Type)
}

func GetStreamer(ctx context.Context, cfg *ProviderConfig, reg Registry, groupMessageHandler GroupMessageHandler) (Streamer, error) {
	lock.Lock()
	defer lock.Unlock()

	if factory, ok := StreamProviders[cfg.Stream.Type]; ok {
		return factory(ctx, cfg, reg, groupMessageHandler)
	}

	return nil, fmt.Errorf("streamer provider not found: %s", cfg.Stream.Type)
}
