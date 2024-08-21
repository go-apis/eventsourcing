package gpub

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/go-apis/eventsourcing/es"
)

type Unsubscribe func(ctx context.Context) error

type streamer struct {
	service string
	client  *pubsub.Client
	topic   *pubsub.Topic

	cctx   context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	errCh chan error

	registeredMu sync.RWMutex
	unsubscribe  []Unsubscribe
}

func (s *streamer) createSubscription(ctx context.Context, suffix string) (*pubsub.Subscription, Unsubscribe, error) {
	subscriptionId := s.topic.ID() + "__" + s.service
	if suffix != "" {
		subscriptionId += "-" + suffix
	}

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

func (s *streamer) loop(sub *pubsub.Subscription, handler es.MessageHandler) {
	defer s.wg.Done()

	h := func(ctx context.Context, msg *pubsub.Message) {
		raw := msg.Data

		if err := handler(s.cctx, raw); err != nil {
			s.errCh <- fmt.Errorf("could not handle message: %w", err)
			msg.Nack()
		}
		msg.Ack()
	}

	for {
		select {
		case <-s.cctx.Done():
			return
		default:
			if err := sub.Receive(s.cctx, h); err != nil {
				s.errCh <- fmt.Errorf("could not receive: %w", err)

				// Retry the receive loop if there was an error.
				time.Sleep(time.Second)
				continue
			}
		}
	}
}

func (s *streamer) AddHandler(ctx context.Context, name string, handler es.MessageHandler) error {
	// Check handler existence.
	s.registeredMu.Lock()
	defer s.registeredMu.Unlock()

	sub, unsubscribe, err := s.createSubscription(ctx, name)
	if err != nil {
		return err
	}

	if unsubscribe != nil {
		s.unsubscribe = append(s.unsubscribe, unsubscribe)
	}

	go s.loop(sub, handler)

	s.wg.Add(1)
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

func NewStreamer(ctx context.Context, service string, config *es.GcpPubSubConfig) (es.Streamer, error) {
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
	s := &streamer{
		service: service,
		client:  client,
		topic:   topic,
		cctx:    cctx,
		cancel:  cancel,
		errCh:   make(chan error, 100),
	}
	return s, nil
}
