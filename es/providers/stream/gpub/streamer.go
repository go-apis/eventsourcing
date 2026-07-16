package gpub

import (
	"context"
	"fmt"
	"os"
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
		Topic:       s.topic,
		AckDeadline: 10 * time.Second,
		// Ordered subscriptions route through the emulator's broken
		// OrderedMessageBacklog even for unkeyed messages — see newTopic.
		EnableMessageOrdering: os.Getenv("PUBSUB_EMULATOR_HOST") == "",
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
			// Nack only: falling through to Ack here would confirm the
			// message and the failed event would never be redelivered.
			msg.Nack()
			return
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
	return publishEvent(ctx, s.topic, evt)
}

func (s *streamer) PublishRaw(ctx context.Context, orderingKey string, payload []byte) error {
	return publishRaw(ctx, s.topic, orderingKey, payload)
}

func (s *streamer) PublishRawBatch(ctx context.Context, msgs []es.RawEvent) []error {
	return publishRawBatch(ctx, s.topic, msgs)
}

func publishEvent(ctx context.Context, topic *pubsub.Topic, evt *es.Event) error {
	data, err := es.MarshalEvent(ctx, evt)
	if err != nil {
		return err
	}

	return publishRaw(ctx, topic, es.EventOrderingKey(evt), data)
}

func publishRaw(ctx context.Context, topic *pubsub.Topic, orderingKey string, payload []byte) error {
	if !topic.EnableMessageOrdering {
		orderingKey = ""
	}
	msg := &pubsub.Message{
		Data:        payload,
		OrderingKey: orderingKey,
	}

	rsp := topic.Publish(ctx, msg)
	if _, err := rsp.Get(ctx); err != nil {
		// A failed publish pauses its ordering key on this topic handle;
		// without a resume every retry of the same key fails immediately
		// for the life of the process.
		topic.ResumePublish(orderingKey)
		return err
	}
	return nil
}

// publishRawBatch fires every message before awaiting any ack, so the
// client batches on the wire (see newTopic's PublishSettings) instead of
// paying a round trip per event; the ordering key still serializes
// same-key messages client-side.
func publishRawBatch(ctx context.Context, topic *pubsub.Topic, msgs []es.RawEvent) []error {
	// The emulator's StreamingPullPusher threads crash under concurrent
	// batched publishes (the v0.5.x sequential trickle never hit this);
	// emulator runs publish one at a time.
	if os.Getenv("PUBSUB_EMULATOR_HOST") != "" {
		errs := make([]error, len(msgs))
		for i, m := range msgs {
			errs[i] = publishRaw(ctx, topic, m.OrderingKey, m.Payload)
		}
		return errs
	}

	results := make([]*pubsub.PublishResult, len(msgs))
	for i, m := range msgs {
		key := m.OrderingKey
		if !topic.EnableMessageOrdering {
			key = ""
		}
		results[i] = topic.Publish(ctx, &pubsub.Message{
			Data:        m.Payload,
			OrderingKey: key,
		})
	}

	errs := make([]error, len(msgs))
	for i, rsp := range results {
		if _, err := rsp.Get(ctx); err != nil {
			topic.ResumePublish(msgs[i].OrderingKey)
			errs[i] = err
		}
	}
	return errs
}

func newTopic(client *pubsub.Client, topicId string) *pubsub.Topic {
	topic := client.Topic(topicId)
	// The Pub/Sub emulator's ordered-message backlog is broken (pulls NPE
	// once keyed messages accumulate — messages become undeliverable), so
	// emulator runs publish unordered; publishRaw/publishRawBatch strip
	// ordering keys to match, as the client requires.
	topic.EnableMessageOrdering = os.Getenv("PUBSUB_EMULATOR_HOST") == ""
	topic.PublishSettings.ByteThreshold = 5000
	topic.PublishSettings.CountThreshold = 10
	topic.PublishSettings.DelayThreshold = 100 * time.Millisecond
	return topic
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

	topic := newTopic(client, config.TopicId)

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
