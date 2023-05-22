package npub

import (
	"context"
	"fmt"

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

	started     bool
	serviceName string
}

func (s *streamer) Start(ctx context.Context, cfg es.Config, callback es.EventCallback) error {
	_, span := otel.Tracer("npub").Start(ctx, "Start")
	defer span.End()

	if cfg == nil {
		return fmt.Errorf("cfg is required")
	}
	if callback == nil {
		return fmt.Errorf("callback is required")
	}

	serviceName := cfg.GetProviderConfig().ServiceName

	handle := func(ctx context.Context, data []byte) error {
		pctx, span := otel.Tracer("npub").Start(ctx, "Handle")
		defer span.End()

		with, err := es.UnmarshalEvent(pctx, cfg, data)
		if err != nil {
			return err
		}
		if with == nil {
			return nil
		}

		return callback(pctx, with.Event)
	}

	_, err := s.conn.QueueSubscribe(s.subject+".*", serviceName, wrapped(handle))
	if err != nil {
		return err
	}

	s.started = true
	s.serviceName = serviceName
	return nil
}

func (s *streamer) Publish(ctx context.Context, evt ...*es.Event) error {
	_, span := otel.Tracer("npub").Start(ctx, "Publish")
	defer span.End()

	if !s.started {
		return fmt.Errorf("streamer is not started")
	}

	datums := make([]Data, len(evt))
	for i, e := range evt {
		data, err := es.MarshalEvent(ctx, s.serviceName, e)
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
