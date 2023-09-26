package noop

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
)

func New(ctx context.Context, cfg es.StreamConfig) (es.Streamer, error) {
	if cfg.Type != "noop" {
		return nil, fmt.Errorf("invalid data provider type: %s", cfg.Type)
	}
	return NewStreamer()
}

func init() {
	es.RegisterStreamProviders("noop", New)
}
