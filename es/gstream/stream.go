package gstream

import (
	"context"
	"fmt"
	"strings"

	"github.com/contextcloud/eventstore/es"
	"go.opentelemetry.io/otel"
)

type streamer struct {
	cli Client
}

func (s *streamer) Start(ctx context.Context, opts es.InitializeOptions, callback es.Callback) error {
	pctx, span := otel.Tracer("gstream").Start(ctx, "Start")
	defer span.End()

	if len(opts.ServiceName) == 0 {
		return fmt.Errorf("service name is required")
	}
	if callback == nil {
		return fmt.Errorf("callback is required")
	}

	mapper := map[string]es.EventDataFunc{}
	for _, eventConfigs := range opts.EventConfigs {
		mapper[eventConfigs.Name] = eventConfigs.Factory
	}

	sub, err := s.cli.Subscription(pctx, opts.ServiceName)
	if err != nil {
		return err
	}

	handle := func(ctx context.Context, data []byte) error {
		pctx, span := otel.Tracer("gstream").Start(ctx, "Handle")
		defer span.End()

		evt, err := UnmarshalEvent(pctx, mapper, data)
		if err != nil {
			return err
		}
		if evt == nil {
			return nil
		}

		// if we are the same service name than do nothing too
		if strings.EqualFold(evt.ServiceName, opts.ServiceName) {
			return nil
		}

		return callback(pctx, evt)
	}

	// is this blocking?
	go func() {
		ctx := context.Background()
		pctx, span := otel.Tracer("gstream").Start(ctx, "Receive")
		defer span.End()

		if err := sub.Receive(pctx, handle); err != nil {
			// no sure what to do here yet
			panic(err)
		}
	}()

	return nil
}

func (s *streamer) Publish(ctx context.Context, evt ...*es.Event) error {
	pctx, span := otel.Tracer("gstream").Start(ctx, "Publish")
	defer span.End()

	datums := make([][]byte, len(evt))
	for i, e := range evt {
		data, err := MarshalEvent(ctx, e)
		if err != nil {
			return err
		}
		datums[i] = data
	}

	for _, data := range datums {
		_, err := s.cli.Publish(pctx, data)
		if err != nil {
			// todo add some logging
			return err
		}
	}

	return nil
}

func (s *streamer) Close(ctx context.Context) error {
	_, span := otel.Tracer("gstream").Start(ctx, "Close")
	defer span.End()

	return s.cli.Close()
}

func NewStreamer(cli Client) (es.Streamer, error) {
	return &streamer{
		cli: cli,
	}, nil
}
