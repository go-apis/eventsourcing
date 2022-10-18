package gpub

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/pkg/gcppubsub"
	"go.opentelemetry.io/otel"
)

func wrapped(callback func(context.Context, []byte) error) func(ctx context.Context, msg *pubsub.Message) {
	return func(ctx context.Context, msg *pubsub.Message) {
		if err := callback(ctx, msg.Data); err != nil {
			msg.Nack()
			return
		}
		msg.Ack()
	}
}

type streamer struct {
	p *gcppubsub.Pub
}

func (s *streamer) Start(ctx context.Context, opts es.InitializeOptions, callback es.EventCallback) error {
	pctx, span := otel.Tracer("gpub").Start(ctx, "Start")
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

	sub, err := s.p.Subscription(pctx, opts.ServiceName)
	if err != nil {
		return err
	}

	handle := func(ctx context.Context, data []byte) error {
		pctx, span := otel.Tracer("gpub").Start(ctx, "Handle")
		defer span.End()

		evt, err := es.UnmarshalEvent(pctx, mapper, data)
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
		pctx, span := otel.Tracer("gpub").Start(ctx, "Receive")
		defer span.End()

		if err := sub.Receive(pctx, wrapped(handle)); err != nil {
			// no sure what to do here yet
			panic(err)
		}
	}()

	return nil
}

func (s *streamer) Publish(ctx context.Context, evt ...*es.Event) error {
	pctx, span := otel.Tracer("gpub").Start(ctx, "Publish")
	defer span.End()

	datums := make([][]byte, len(evt))
	for i, e := range evt {
		data, err := es.MarshalEvent(ctx, e)
		if err != nil {
			return err
		}
		datums[i] = data
	}

	for _, data := range datums {
		_, err := s.p.Publish(pctx, data)
		if err != nil {
			// todo add some logging
			return err
		}
	}

	return nil
}

func (s *streamer) Close(ctx context.Context) error {
	_, span := otel.Tracer("gpub").Start(ctx, "Close")
	defer span.End()

	return s.p.Close()
}

func NewStreamer(p *gcppubsub.Pub) (es.Streamer, error) {
	return &streamer{
		p: p,
	}, nil
}
