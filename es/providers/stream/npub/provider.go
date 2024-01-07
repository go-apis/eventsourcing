package npub

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
)

func New(ctx context.Context, cfg *es.ProviderConfig, reg es.Registry) (es.Streamer, error) {
	if cfg.Stream.Type != "nats" {
		return nil, fmt.Errorf("invalid data provider type: %s", cfg.Stream.Type)
	}
	if cfg.Stream.Nats == nil {
		return nil, fmt.Errorf("invalid nats config")
	}
	if cfg.Stream.Nats.Url == "" {
		return nil, fmt.Errorf("invalid nats url")
	}
	if cfg.Stream.Nats.Subject == "" {
		return nil, fmt.Errorf("invalid nats subject")
	}

	return NewStreamer(ctx, cfg.Service, cfg.Stream.Nats, reg.ParseEvent)
}

func init() {
	es.RegisterStreamProviders("nats", New)
}
