package gpub

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/contextcloud/eventstore/es"
	"github.com/google/uuid"
)

type Unsubscribe func(ctx context.Context) error

type streamer struct {
	service             string
	registry            es.Registry
	groupMessageHandler es.GroupMessageHandler
	client              *pubsub.Client
	topic               *pubsub.Topic

	cctx   context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	errCh chan error

	registered   map[string]bool
	registeredMu sync.RWMutex
	unsubscribe  []Unsubscribe
}

func (s *streamer) createSubscription(ctx context.Context, suffix string) (*pubsub.Subscription, Unsubscribe, error) {
	subscriptionId := s.service
	if suffix != "" {
		subscriptionId = fmt.Sprintf("%s-%s", s.service, suffix)
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

func (s *streamer) loop(sub *pubsub.Subscription, group string) {
	defer s.wg.Done()

	for {
		if err := sub.Receive(s.cctx, s.handle(group)); err != nil {
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

func (s *streamer) handle(group string) func(context.Context, *pubsub.Message) {
	return func(ctx context.Context, msg *pubsub.Message) {
		if err := s.groupMessageHandler.HandleGroupMessage(ctx, group, msg.Data); err != nil {
			select {
			case s.errCh <- err:
			default:
				log.Printf("missed error in GCP event bus: %s", err)
			}
			msg.Nack()
			return
		}
		msg.Ack()
	}
}

func (s *streamer) addGroup(ctx context.Context, group string) error {
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

	sub, unsubscribe, err := s.createSubscription(ctx, suffix)
	if err != nil {
		return err
	}

	if unsubscribe != nil {
		s.unsubscribe = append(s.unsubscribe, unsubscribe)
	}

	// Register handler.
	s.registered[group] = true
	s.wg.Add(1)

	go s.loop(sub, group)
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

func NewStreamer(ctx context.Context, service string, config *es.GcpPubSubConfig, reg es.Registry, groupMessageHandler es.GroupMessageHandler) (es.Streamer, error) {
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
		service:             service,
		registry:            reg,
		groupMessageHandler: groupMessageHandler,
		client:              client,
		topic:               topic,
		cctx:                cctx,
		cancel:              cancel,
		registered:          make(map[string]bool),
		errCh:               make(chan error, 100),
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
