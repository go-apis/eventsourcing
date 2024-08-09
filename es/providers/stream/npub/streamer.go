package npub

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-apis/eventsourcing/es"
	"github.com/google/uuid"

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
	service             string
	registry            es.Registry
	groupMessageHandler es.GroupMessageHandler
	streamName          string
	conn                *nats.Conn
	js                  nats.JetStreamContext
	stream              *nats.StreamInfo

	cctx   context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	errCh chan error

	registered   map[string]bool
	registeredMu sync.RWMutex
	unsubscribe  []Unsubscribe
}

func (s *streamer) createQueueSubscriber(subject string, group string, suffix string) (*nats.Subscription, Unsubscribe, error) {
	consumerName := s.service
	if suffix != "" {
		consumerName = fmt.Sprintf("%s-%s", s.service, suffix)
	}

	sub, err := s.js.QueueSubscribe(subject, consumerName, s.handle(group),
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

func (s *streamer) handle(group string) nats.MsgHandler {
	return func(msg *nats.Msg) {
		if err := s.groupMessageHandler.HandleGroupMessage(s.cctx, group, msg.Data); err != nil {
			select {
			case s.errCh <- err:
			default:
				log.Printf("missed error in NATS event bus: %s", err)
			}
			msg.Nak()
			return
		}

		msg.AckSync()
	}
}

func (s *streamer) addGroup(ctx context.Context, group string) error {
	_, span := otel.Tracer("noop").Start(ctx, "AddHandler")
	defer span.End()

	// Check handler existence.
	s.registeredMu.Lock()
	defer s.registeredMu.Unlock()

	if _, ok := s.registered[group]; ok {
		return fmt.Errorf("handler already registered: %s", group)
	}

	suffix := ""
	switch group {
	case es.InternalGroup:
		return fmt.Errorf("invalid group name: %s", group)
	case es.ExternalGroup:
		suffix = ""
	case es.RandomGroup:
		suffix = uuid.NewString()
	default:
		suffix = group
	}

	subject := fmt.Sprintf("%s.*.*", s.streamName)
	sub, unsubscribe, err := s.createQueueSubscriber(subject, group, suffix)
	if err != nil {
		return err
	}

	if unsubscribe != nil {
		s.unsubscribe = append(s.unsubscribe, unsubscribe)
	}

	s.registered[group] = true
	s.wg.Add(1)

	go s.loop(sub)

	return nil
}

func (s *streamer) Publish(ctx context.Context, evt *es.Event) error {
	_, span := otel.Tracer("npub").Start(ctx, "Publish")
	defer span.End()

	data, err := es.MarshalEvent(ctx, evt)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("%s.%s.%s", s.streamName, s.service, evt.Type)
	if err := s.conn.Publish(subject, data); err != nil {
		return err
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

func NewStreamer(ctx context.Context, service string, natsConfig *es.NatsConfig, reg es.Registry, groupMessageHandler es.GroupMessageHandler) (es.Streamer, error) {
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
	s := &streamer{
		service:             service,
		registry:            reg,
		groupMessageHandler: groupMessageHandler,
		streamName:          natsConfig.Subject,
		conn:                conn,
		js:                  js,
		stream:              streamInfo,

		cctx:       cctx,
		cancel:     cancel,
		registered: make(map[string]bool),
		errCh:      make(chan error, 100),
	}

	for _, group := range reg.GetGroups() {
		if group == es.InternalGroup {
			continue
		}

		if err := s.addGroup(ctx, group); err != nil {
			return nil, err
		}
	}
	return s, nil
}
