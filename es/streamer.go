package es

import (
	"context"
	"net/http"
)

type StreamerFactory func(ctx context.Context, cfg *ProviderConfig) (Streamer, error)

// PushReceiver is implemented by streamers that consume deliveries over
// HTTP (e.g. Pub/Sub push subscriptions) instead of in-process pull loops.
// The returned handler acknowledges a delivery with 2xx and rejects it for
// redelivery with 5xx.
type PushReceiver interface {
	PushHandler() http.Handler
}

// PushHandlerProvider is implemented by Client when the configured streamer
// consumes over HTTP; mount the returned handler on the service's router.
type PushHandlerProvider interface {
	PushHandler() (http.Handler, bool)
}

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
