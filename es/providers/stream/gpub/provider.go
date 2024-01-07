package gpub

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
)

func New(ctx context.Context, cfg *es.ProviderConfig, reg es.Registry) (es.Streamer, error) {
	if cfg.Stream.Type != "pubsub" {
		return nil, fmt.Errorf("invalid data provider type: %s", cfg.Stream.Type)
	}
	if cfg.Stream.PubSub == nil {
		return nil, fmt.Errorf("invalid pubsub config")
	}

	return NewStreamer(ctx, cfg.Service, cfg.Stream.PubSub, reg.ParseEvent)
}

func init() {
	es.RegisterStreamProviders("pubsub", New)
}
