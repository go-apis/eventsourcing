package es

import (
	"context"
)

type StreamerFactory func(ctx context.Context, cfg *ProviderConfig) (Streamer, error)

type Streamer interface {
	AddHandler(ctx context.Context, name string, handler EventHandler) error
	Publish(ctx context.Context, evt ...*Event) error
	Errors() <-chan error
	Close(ctx context.Context) error
}
