package gpub

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/pkg/gcppubsub"
)

func New(ctx context.Context, cfg *es.ProviderConfig) (es.Streamer, error) {
	if cfg.Stream.Type != "pubsub" {
		return nil, fmt.Errorf("invalid data provider type: %s", cfg.Stream.Type)
	}
	if cfg.Stream.PubSub == nil {
		return nil, fmt.Errorf("invalid pubsub config")
	}

	// create a new gorm connection
	p, err := gcppubsub.Open(cfg.Stream.PubSub)
	if err != nil {
		return nil, err
	}

	return NewStreamer(cfg.Service, p)
}

func init() {
	es.RegisterStreamProviders("pubsub", New)
}
