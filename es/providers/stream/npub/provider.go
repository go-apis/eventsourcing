package npub

import (
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/nats-io/nats.go"
)

func New(cfg es.StreamConfig) (es.Streamer, error) {
	if cfg.Type != "nats" {
		return nil, fmt.Errorf("invalid data provider type: %s", cfg.Type)
	}
	if cfg.Nats == nil {
		return nil, fmt.Errorf("invalid nats config")
	}
	if cfg.Nats.Url == "" {
		return nil, fmt.Errorf("invalid nats url")
	}
	if cfg.Nats.Subject == "" {
		return nil, fmt.Errorf("invalid nats subject")
	}

	conn, err := nats.Connect(cfg.Nats.Url)
	if err != nil {
		return nil, err
	}

	return NewStreamer(conn, cfg.Nats.Subject)
}

func init() {
	es.RegisterStreamProviders("nats", New)
}
