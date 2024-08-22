package mpub

import (
	"context"
	"fmt"
	"sync"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-apis/eventsourcing/es"

	"go.opentelemetry.io/otel"
)

type streamer struct {
	cctx   context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	topic  string
	pubsub es.MemoryBusPubSub

	errCh chan error

	registeredMu sync.RWMutex
}

func (s *streamer) loop(messages <-chan *message.Message, handler es.MessageHandler) {
	defer s.wg.Done()

	h := func(msg *message.Message) error {
		raw := msg.Payload

		if err := handler(s.cctx, raw); err != nil {
			msg.Nack()
			return fmt.Errorf("could not handle message: %w", err)
		}
		if ok := msg.Ack(); !ok {
			return fmt.Errorf("failed to ack message")
		}
		return nil
	}

	for {
		select {
		case <-s.cctx.Done():
			return
		default:
			msg, ok := <-messages
			if !ok {
				return
			}
			if err := h(msg); err != nil {
				s.errCh <- err
			}
		}
	}
}

func (s *streamer) AddHandler(ctx context.Context, name string, handler es.MessageHandler) error {
	s.registeredMu.Lock()
	defer s.registeredMu.Unlock()

	messages, err := s.pubsub.Subscribe(s.cctx, s.topic)
	if err != nil {
		return err
	}

	go s.loop(messages, handler)

	s.wg.Add(1)
	return nil
}

func (s *streamer) Publish(ctx context.Context, evt *es.Event) error {
	_, span := otel.Tracer("mpub").Start(ctx, "Publish")
	defer span.End()

	key := fmt.Sprintf("%s:%s:%s:%d", evt.Namespace, evt.AggregateId.String(), evt.AggregateType, evt.Version)
	data, err := es.MarshalEvent(ctx, evt)
	if err != nil {
		return err
	}

	msg := message.NewMessage(key, data)
	if err := s.pubsub.Publish(s.topic, msg); err != nil {
		return err
	}
	return nil
}

func (s *streamer) Errors() <-chan error {
	return s.errCh
}

func (s *streamer) Close(ctx context.Context) error {
	_, span := otel.Tracer("mpub").Start(ctx, "Close")
	defer span.End()

	s.cancel()
	s.wg.Wait()

	return nil
}

func NewStreamer(ctx context.Context, topic string, pubsub es.MemoryBusPubSub) (es.Streamer, error) {
	if topic == "" {
		return nil, fmt.Errorf("missing topic")
	}
	if pubsub == nil {
		return nil, fmt.Errorf("missing pubsub")
	}

	cctx, cancel := context.WithCancel(ctx)
	s := &streamer{
		cctx:   cctx,
		cancel: cancel,
		topic:  topic,
		pubsub: pubsub,
		errCh:  make(chan error, 100),
	}
	return s, nil
}
