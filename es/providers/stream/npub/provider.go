package npub

import (
	"context"
	"fmt"

	"github.com/go-apis/eventsourcing/es"
)

func New(ctx context.Context, cfg *es.ProviderConfig, reg es.Registry, groupMessageHandler es.GroupMessageHandler) (es.Streamer, error) {
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

	return NewStreamer(ctx, cfg.Service, cfg.Stream.Nats, reg, groupMessageHandler)
}

func init() {
	es.RegisterStreamProviders("nats", New)
}
