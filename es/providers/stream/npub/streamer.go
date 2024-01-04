package npub

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/contextcloud/eventstore/es"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
)

type Unsubscribe func(ctx context.Context) error

func GetOrCreateStream(js nats.JetStreamContext, streamName string) (*nats.StreamInfo, error) {
	if info, err := js.StreamInfo(streamName); err == nil {
		return info, nil
	}

	streamConfig := &nats.StreamConfig{
		Name:      streamName,
		Subjects:  []string{streamName + ".*.*"},
		Storage:   nats.FileStorage,
		Retention: nats.InterestPolicy,
	}
	return js.AddStream(streamConfig)
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
	service    string
	streamName string
	conn       *nats.Conn
	js         nats.JetStreamContext
	stream     *nats.StreamInfo

	cctx   context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	errCh chan error

	registered   map[string]bool
	registeredMu sync.RWMutex
	unsubscribe  []Unsubscribe
}

func (s *streamer) createQueueSubscriber(subject string, consumerName string, handler es.EventHandler) (*nats.Subscription, Unsubscribe, error) {
	sub, err := s.js.QueueSubscribe(subject, consumerName, s.handle(handler),
		nats.DeliverNew(),
		nats.ManualAck(),
		nats.AckExplicit(),
		nats.AckWait(60*time.Second),
		nats.MaxDeliver(10),
	)
	if err != nil {
		return nil, nil, err
	}

	// TODO we need to figure out how to unsubscribe from this.

	return sub, nil, nil
}

// Handles all events coming in on the channel.
func (s *streamer) loop(sub *nats.Subscription) {
	defer s.wg.Done()

	for {
		select {
		case <-s.cctx.Done():
			if s.cctx.Err() != context.Canceled {
				log.Printf("context error in NATS event bus: %s", s.cctx.Err())
			}

			return
		}
	}
}

func (s *streamer) handle(handler es.EventHandler) nats.MsgHandler {
	return func(msg *nats.Msg) {
		evt, err := es.GlobalRegistry.ParseEvent(s.cctx, msg.Data)
		if err != nil && !errors.Is(err, es.ErrNotFound) {
			select {
			case s.errCh <- err:
			default:
				log.Printf("missed error in NATS event bus: %s", err)
			}
			msg.Nak()
			return
		}

		if evt != nil {
			if err := handler.Handle(s.cctx, evt); err != nil {
				select {
				case s.errCh <- err:
				default:
					log.Printf("missed error in NATS event bus: %s", err)
				}
				msg.Nak()
				return
			}
		}

		msg.AckSync()
	}
}

func (s *streamer) AddHandler(ctx context.Context, name string, handler es.EventHandler) error {
	_, span := otel.Tracer("noop").Start(ctx, "AddHandler")
	defer span.End()

	// Check handler existence.
	s.registeredMu.Lock()
	defer s.registeredMu.Unlock()

	if _, ok := s.registered[name]; ok {
		return fmt.Errorf("handler already registered: %s", name)
	}

	subject := fmt.Sprintf("%s.*.*", s.streamName)
	consumerName := s.service
	if name != "" {
		consumerName = fmt.Sprintf("%s-%s", s.service, name)
	}

	sub, unsubscribe, err := s.createQueueSubscriber(subject, consumerName, handler)
	if err != nil {
		return err
	}

	if unsubscribe != nil {
		s.unsubscribe = append(s.unsubscribe, unsubscribe)
	}

	s.registered[name] = true
	s.wg.Add(1)

	go s.loop(sub)

	return nil
}

func (s *streamer) Publish(ctx context.Context, evt ...*es.Event) error {
	_, span := otel.Tracer("npub").Start(ctx, "Publish")
	defer span.End()

	for _, e := range evt {
		data, err := es.MarshalEvent(ctx, e)
		if err != nil {
			return err
		}

		subject := fmt.Sprintf("%s.%s.%s", s.streamName, s.service, e.Type)
		if err := s.conn.Publish(subject, data); err != nil {
			return err
		}
	}

	return nil
}

func (s *streamer) Errors() <-chan error {
	return s.errCh
}

func (s *streamer) Close(ctx context.Context) error {
	_, span := otel.Tracer("npub").Start(ctx, "Close")
	defer span.End()

	s.cancel()
	s.wg.Wait()

	// unsubscribe any ephemeral subscribers we created.
	for _, unsub := range s.unsubscribe {
		if err := unsub(ctx); err != nil {
			s.errCh <- err
		}
	}

	s.conn.Close()
	return s.conn.LastError()
}

func NewStreamer(ctx context.Context, service string, natsConfig *es.NatsConfig) (es.Streamer, error) {
	conn, err := nats.Connect(natsConfig.Url)
	if err != nil {
		return nil, err
	}

	js, err := conn.JetStream()
	if err != nil {
		return nil, err
	}

	streamInfo, err := GetOrCreateStream(js, natsConfig.Subject)
	if err != nil {
		return nil, err
	}

	cctx, cancel := context.WithCancel(ctx)
	return &streamer{
		service:    service,
		streamName: natsConfig.Subject,
		conn:       conn,
		js:         js,
		stream:     streamInfo,

		cctx:       cctx,
		cancel:     cancel,
		registered: make(map[string]bool),
		errCh:      make(chan error, 100),
	}, nil
}
