package noop

import (
	"context"

	"github.com/contextcloud/eventstore/es"

	"go.opentelemetry.io/otel"
)

type streamer struct {
}

func (s *streamer) Start(ctx context.Context, cfg es.Config, callback es.EventCallback) error {
	_, span := otel.Tracer("noop").Start(ctx, "Start")
	defer span.End()

	return nil
}

func (s *streamer) Publish(ctx context.Context, evt ...*es.Event) error {
	_, span := otel.Tracer("noop").Start(ctx, "Publish")
	defer span.End()

	return nil
}

func (s *streamer) Close(ctx context.Context) error {
	_, span := otel.Tracer("noop").Start(ctx, "Close")
	defer span.End()

	return nil
}

func NewStreamer() (es.Streamer, error) {
	return &streamer{}, nil
}
