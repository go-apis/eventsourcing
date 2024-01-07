package gpub

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/contextcloud/eventstore/es"
)

type Unsubscribe func(ctx context.Context) error

type streamer struct {
	service string
	parser  es.EventParser
	client  *pubsub.Client
	topic   *pubsub.Topic

	cctx   context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	errCh chan error

	registered   map[string]bool
	registeredMu sync.RWMutex
	unsubscribe  []Unsubscribe
}

func (s *streamer) createSubscription(ctx context.Context, subscriptionId string) (*pubsub.Subscription, Unsubscribe, error) {
	existingSub := s.client.Subscription(subscriptionId)
	exists, err := existingSub.Exists(ctx)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		existingSub.ReceiveSettings.MaxOutstandingMessages = 100
		return existingSub, nil, nil
	}

	sub, err := s.client.CreateSubscription(ctx, subscriptionId, pubsub.SubscriptionConfig{
		Topic:                 s.topic,
		AckDeadline:           10 * time.Second,
		EnableMessageOrdering: true,
		RetryPolicy: &pubsub.RetryPolicy{
			MinimumBackoff: 10 * time.Millisecond,
		},
	})
	if err != nil {
		return nil, nil, err
	}

	sub.ReceiveSettings.MaxOutstandingMessages = 100
	return sub, sub.Delete, nil
}

func (s *streamer) loop(sub *pubsub.Subscription, handler es.EventHandler) {
	defer s.wg.Done()

	for {
		if err := sub.Receive(s.cctx, s.handle(handler)); err != nil {
			err = fmt.Errorf("could not receive: %w", err)

			select {
			case s.errCh <- err:
			default:
				log.Printf("missed error in GCP event bus: %s", err)
			}

			// Retry the receive loop if there was an error.
			time.Sleep(time.Second)
			continue
		}

		return
	}
}

func (s *streamer) handle(handler es.EventHandler) func(context.Context, *pubsub.Message) {
	return func(ctx context.Context, msg *pubsub.Message) {
		evt, err := s.parser(ctx, msg.Data)
		if err != nil && !errors.Is(err, es.ErrNotFound) {
			select {
			case s.errCh <- err:
			default:
				log.Printf("missed error in GCP event bus: %s", err)
			}
			msg.Nack()
			return
		}

		if evt != nil {
			if err := handler.Handle(ctx, evt); err != nil {
				select {
				case s.errCh <- err:
				default:
					log.Printf("missed error in GCP event bus: %s", err)
				}
				msg.Nack()
				return
			}
		}

		msg.Ack()
	}
}

func (s *streamer) AddHandler(ctx context.Context, name string, handler es.EventHandler) error {
	// Check handler existence.
	s.registeredMu.Lock()
	defer s.registeredMu.Unlock()

	if _, ok := s.registered[name]; ok {
		return fmt.Errorf("handler already registered: %s", name)
	}

	subscriptionId := s.service
	if name != "" {
		subscriptionId += "-" + name
	}
	sub, unsubscribe, err := s.createSubscription(ctx, subscriptionId)
	if err != nil {
		return err
	}

	if unsubscribe != nil {
		s.unsubscribe = append(s.unsubscribe, unsubscribe)
	}

	// Register handler.
	s.registered[name] = true
	s.wg.Add(1)

	go s.loop(sub, handler)
	return nil
}

func (s *streamer) Publish(ctx context.Context, evt *es.Event) error {
	orderingKey := fmt.Sprintf("%s:%s:%s:%d", evt.Namespace, evt.AggregateId.String(), evt.AggregateType, evt.Version)
	data, err := es.MarshalEvent(ctx, evt)
	if err != nil {
		return err
	}

	msg := &pubsub.Message{
		Data:        data,
		OrderingKey: orderingKey,
	}

	rsp := s.topic.Publish(ctx, msg)
	if _, err := rsp.Get(ctx); err != nil {
		return err
	}
	return nil
}

func (s *streamer) Errors() <-chan error {
	return s.errCh
}

func (s *streamer) Close(ctx context.Context) error {
	s.topic.Stop()

	s.cancel()
	s.wg.Wait()

	// unsubscribe any ephemeral subscribers we created.
	for _, unsub := range s.unsubscribe {
		if err := unsub(ctx); err != nil {
			s.errCh <- err
		}
	}

	return s.client.Close()
}

func NewStreamer(ctx context.Context, service string, config *es.GcpPubSubConfig, parser es.EventParser) (es.Streamer, error) {
	client, err := pubsub.NewClient(ctx, config.ProjectId)
	if err != nil {
		return nil, err
	}

	topic := client.Topic(config.TopicId)
	topic.EnableMessageOrdering = true
	topic.PublishSettings.ByteThreshold = 5000
	topic.PublishSettings.CountThreshold = 10
	topic.PublishSettings.DelayThreshold = 100 * time.Millisecond

	cctx, cancel := context.WithCancel(ctx)
	return &streamer{
		service:    service,
		parser:     parser,
		client:     client,
		topic:      topic,
		cctx:       cctx,
		cancel:     cancel,
		registered: make(map[string]bool),
		errCh:      make(chan error, 100),
	}, nil
}
