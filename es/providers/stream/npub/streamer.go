package npub

import (
	"context"
	"fmt"
	"strings"

	"github.com/contextcloud/eventstore/es"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
)

type Data struct {
	Name string
	Raw  []byte
}

func wrapped(callback func(context.Context, []byte) error) func(msg *nats.Msg) {
	return func(msg *nats.Msg) {
		ctx := context.Background()

		if err := callback(ctx, msg.Data); err != nil {
			msg.Nak()
			return
		}
		msg.Ack()
	}
}

type streamer struct {
	conn    *nats.Conn
	subject string
}

func (s *streamer) Start(ctx context.Context, opts es.InitializeOptions, callback es.EventCallback) error {
	_, span := otel.Tracer("npub").Start(ctx, "Start")
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

	handle := func(ctx context.Context, data []byte) error {
		pctx, span := otel.Tracer("npub").Start(ctx, "Handle")
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

	_, err := s.conn.QueueSubscribe(s.subject+".*", opts.ServiceName, wrapped(handle))
	if err != nil {
		return err
	}
	return nil
}

func (s *streamer) Publish(ctx context.Context, evt ...*es.Event) error {
	_, span := otel.Tracer("npub").Start(ctx, "Publish")
	defer span.End()

	datums := make([]Data, len(evt))
	for i, e := range evt {
		data, err := es.MarshalEvent(ctx, e)
		if err != nil {
			return err
		}
		datums[i] = Data{
			Name: e.Type,
			Raw:  data,
		}
	}

	for _, d := range datums {
		subject := s.subject + "." + d.Name
		if err := s.conn.Publish(subject, d.Raw); err != nil {
			return err
		}
	}

	return nil
}

func (s *streamer) Close(ctx context.Context) error {
	_, span := otel.Tracer("npub").Start(ctx, "Close")
	defer span.End()

	s.conn.Close()
	return s.conn.LastError()
}

func NewStreamer(conn *nats.Conn, subject string) (es.Streamer, error) {
	return &streamer{
		conn:    conn,
		subject: subject,
	}, nil
}
