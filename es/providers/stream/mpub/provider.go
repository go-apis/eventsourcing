package mpub

import (
	"context"
	"fmt"

	"github.com/go-apis/eventsourcing/es"
)

func New(ctx context.Context, cfg *es.ProviderConfig) (es.Streamer, error) {
	if cfg.Stream.Type != "mpub" {
		return nil, fmt.Errorf("invalid data provider type: %s", cfg.Stream.Type)
	}
	if cfg.Stream.Memory == nil {
		return nil, fmt.Errorf("missing memory pubsub config")
	}
	if cfg.Stream.Memory.PubSub == nil {
		return nil, fmt.Errorf("missing memory pubsub")
	}

	return NewStreamer(ctx, cfg.Stream.Memory.Topic, cfg.Stream.Memory.PubSub)
}

func init() {
	es.RegisterStreamProviders("mpub", New)
}
