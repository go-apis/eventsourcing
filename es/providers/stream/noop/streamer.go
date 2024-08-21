package noop

import (
	"context"

	"github.com/go-apis/eventsourcing/es"

	"go.opentelemetry.io/otel"
)

type streamer struct {
	errCh chan error
}

func (s *streamer) AddHandler(ctx context.Context, name string, handler es.MessageHandler) error {
	_, span := otel.Tracer("noop").Start(ctx, "AddHandler")
	defer span.End()

	return nil
}

func (s *streamer) Publish(ctx context.Context, evt *es.Event) error {
	_, span := otel.Tracer("noop").Start(ctx, "Publish")
	defer span.End()

	return nil
}

func (s *streamer) Errors() <-chan error {
	return s.errCh
}

func (s *streamer) Close(ctx context.Context) error {
	_, span := otel.Tracer("noop").Start(ctx, "Close")
	defer span.End()

	return nil
}

func NewStreamer() (es.Streamer, error) {
	return &streamer{
		errCh: make(chan error, 100),
	}, nil
}
