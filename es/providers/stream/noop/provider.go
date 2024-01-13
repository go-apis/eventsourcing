package noop

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
)

func New(ctx context.Context, cfg *es.ProviderConfig, reg es.Registry, groupMessageHandler es.GroupMessageHandler) (es.Streamer, error) {
	if cfg.Stream.Type != "noop" {
		return nil, fmt.Errorf("invalid data provider type: %s", cfg.Stream.Type)
	}
	return NewStreamer()
}

func init() {
	es.RegisterStreamProviders("noop", New)
}
