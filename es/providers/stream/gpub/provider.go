package gpub

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/pkg/gcppubsub"
)

func New(ctx context.Context, cfg es.StreamConfig) (es.Streamer, error) {
	if cfg.Type != "pubsub" {
		return nil, fmt.Errorf("invalid data provider type: %s", cfg.Type)
	}
	if cfg.PubSub == nil {
		return nil, fmt.Errorf("invalid pubsub config")
	}

	// create a new gorm connection
	p, err := gcppubsub.Open(cfg.PubSub)
	if err != nil {
		return nil, err
	}

	return NewStreamer(p)
}

func init() {
	es.RegisterStreamProviders("pubsub", New)
}
