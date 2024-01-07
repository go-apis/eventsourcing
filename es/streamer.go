package es

import (
	"context"
)

type EventParser func(ctx context.Context, msg []byte) (*Event, error)

type StreamerFactory func(ctx context.Context, cfg *ProviderConfig, reg Registry) (Streamer, error)

type EventPublisher interface {
	Publish(ctx context.Context, evt *Event) error
}

type Streamer interface {
	AddHandler(ctx context.Context, name string, handler EventHandler) error
	Publish(ctx context.Context, evt *Event) error
	Errors() <-chan error
	Close(ctx context.Context) error
}
