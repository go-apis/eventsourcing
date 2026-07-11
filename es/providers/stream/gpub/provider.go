package gpub

import (
	"context"
	"fmt"

	"github.com/go-apis/eventsourcing/es"
)

func New(ctx context.Context, cfg *es.ProviderConfig) (es.Streamer, error) {
	if cfg.Stream.Type != "pubsub" {
		return nil, fmt.Errorf("invalid data provider type: %s", cfg.Stream.Type)
	}
	if cfg.Stream.PubSub == nil {
		return nil, fmt.Errorf("invalid pubsub config")
	}

	switch cfg.Stream.PubSub.Mode {
	case "", "pull":
		return NewStreamer(ctx, cfg.Service, cfg.Stream.PubSub)
	case "push":
		return NewPushStreamer(ctx, cfg.Service, cfg.Stream.PubSub)
	default:
		return nil, fmt.Errorf("invalid pubsub mode: %s", cfg.Stream.PubSub.Mode)
	}
}

func init() {
	es.RegisterStreamProviders("pubsub", New)
}
