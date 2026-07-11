package gpub

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"sync"

	//nolint:staticcheck // v1 client kept for parity with the pull streamer; v2 migration is a separate change.
	"cloud.google.com/go/pubsub"
	"github.com/go-apis/eventsourcing/es"
)

// pushEnvelope is the JSON body Pub/Sub sends to push endpoints.
// Message.Data is base64 in the wire format; encoding/json decodes it
// into []byte directly.
type pushEnvelope struct {
	Message struct {
		Data      []byte `json:"data"`
		MessageId string `json:"messageId"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

// pushStreamer serves Pub/Sub push deliveries over HTTP instead of running
// pull loops. Subscriptions are provisioned externally (Terraform); handlers
// registered via AddHandler are routed by subscription id.
type pushStreamer struct {
	service string
	topicId string
	client  *pubsub.Client
	topic   *pubsub.Topic

	errCh chan error

	mu       sync.RWMutex
	handlers map[string]es.MessageHandler
}

func (s *pushStreamer) subscriptionId(name string) string {
	id := s.topicId + "__" + s.service
	if name != "" {
		id += "-" + name
	}
	return id
}

func (s *pushStreamer) AddHandler(ctx context.Context, name string, handler es.MessageHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.subscriptionId(name)
	if _, exists := s.handlers[id]; exists {
		return fmt.Errorf("handler already registered for subscription %q", id)
	}
	s.handlers[id] = handler
	return nil
}

func (s *pushStreamer) PushHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var env pushEnvelope
		if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
			// Malformed body will never parse on redelivery either.
			http.Error(w, "bad envelope", http.StatusBadRequest)
			return
		}

		s.mu.RLock()
		handler, ok := s.handlers[path.Base(env.Subscription)]
		s.mu.RUnlock()
		if !ok {
			// A subscription pointed at this endpoint that we have no
			// handler for is a provisioning error; 404 keeps it visible in
			// push metrics rather than silently acking.
			http.Error(w, "unknown subscription", http.StatusNotFound)
			return
		}

		if err := handler(r.Context(), env.Message.Data); err != nil {
			slog.ErrorContext(r.Context(), "pubsub push delivery failed",
				"subscription", env.Subscription,
				"messageId", env.Message.MessageId,
				"error", err,
			)
			select {
			case s.errCh <- fmt.Errorf("could not handle message %s: %w", env.Message.MessageId, err):
			default:
			}
			// Non-2xx nacks the delivery so Pub/Sub redelivers with backoff.
			http.Error(w, "handler error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

func (s *pushStreamer) Publish(ctx context.Context, evt *es.Event) error {
	return publishEvent(ctx, s.topic, evt)
}

func (s *pushStreamer) Errors() <-chan error {
	return s.errCh
}

func (s *pushStreamer) Close(ctx context.Context) error {
	s.topic.Stop()
	return s.client.Close()
}

func NewPushStreamer(ctx context.Context, service string, config *es.GcpPubSubConfig) (es.Streamer, error) {
	client, err := pubsub.NewClient(ctx, config.ProjectId)
	if err != nil {
		return nil, err
	}

	return &pushStreamer{
		service:  service,
		topicId:  config.TopicId,
		client:   client,
		topic:    newTopic(client, config.TopicId),
		errCh:    make(chan error, 100),
		handlers: map[string]es.MessageHandler{},
	}, nil
}
