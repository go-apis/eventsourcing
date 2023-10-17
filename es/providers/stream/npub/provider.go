package npub

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/nats-io/nats.go"
)

func New(ctx context.Context, cfg *es.ProviderConfig) (es.Streamer, error) {
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

	conn, err := nats.Connect(cfg.Stream.Nats.Url)
	if err != nil {
		return nil, err
	}

	return NewStreamer(cfg.Service, conn, cfg.Stream.Nats.Subject)
}

func init() {
	es.RegisterStreamProviders("nats", New)
}
