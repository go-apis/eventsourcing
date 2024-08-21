package es

import (
	"context"
)

type StreamerFactory func(ctx context.Context, cfg *ProviderConfig) (Streamer, error)

type EventPublisher interface {
	Publish(ctx context.Context, evt *Event) error
}

type MessageHandler func(ctx context.Context, payload []byte) error

type Streamer interface {
	AddHandler(ctx context.Context, name string, handler MessageHandler) error
	Publish(ctx context.Context, evt *Event) error
	Errors() <-chan error
	Close(ctx context.Context) error
}
