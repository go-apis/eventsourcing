package apub

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
)

func New(ctx context.Context, cfg *es.ProviderConfig, reg es.Registry, groupMessageHandler es.GroupMessageHandler) (es.Streamer, error) {
	if cfg.Stream.Type != "apub" {
		return nil, fmt.Errorf("invalid data provider type: %s", cfg.Stream.Type)
	}
	if cfg.Stream.AWS == nil {
		return nil, fmt.Errorf("invalid aws config")
	}
	if cfg.Stream.AWS.TopicArn == "" {
		return nil, fmt.Errorf("invalid aws topic arn")
	}

	return NewStreamer(ctx, cfg.Service, cfg.Stream.AWS, reg, groupMessageHandler)
}

func init() {
	es.RegisterStreamProviders("apub", New)
}
