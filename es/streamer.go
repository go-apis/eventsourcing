package es

import (
	"context"
)

type StreamerFactory func(ctx context.Context, cfg *ProviderConfig, reg Registry, groupMessageHandler GroupMessageHandler) (Streamer, error)

type EventPublisher interface {
	Publish(ctx context.Context, evt *Event) error
}

type Streamer interface {
	Publish(ctx context.Context, evt *Event) error
	Errors() <-chan error
	Close(ctx context.Context) error
}
