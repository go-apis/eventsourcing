package pb

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"reflect"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/pb/store"
	"github.com/google/uuid"
)

type Streamer interface {
	Run(ctx context.Context) error
	Close(ctx context.Context) error
}

type streamer struct {
	storeClient store.StoreClient
	publisher   es.Publisher
	eventTypes  []string
	eventMap    map[string]reflect.Type
	serviceName string
	cancel      context.CancelFunc
}

func (s *streamer) handle(ctx context.Context, msg *store.EventStreamResponse) error {
	evts := make([]es.Event, len(msg.Events))
	for i, evt := range msg.Events {
		aggregateId, err := uuid.Parse(evt.AggregateId)
		if err != nil {
			return err
		}

		// create the event
		event := es.Event{
			ServiceName:   evt.ServiceName,
			Namespace:     evt.Namespace,
			AggregateId:   aggregateId,
			AggregateType: evt.AggregateType,
			Type:          evt.Type,
			Timestamp:     evt.Timestamp.AsTime(),
		}

		if evt.Data != nil {
			t := s.eventMap[evt.Type]
			event.Data = reflect.New(t).Interface()
			if err := json.Unmarshal(evt.Data, event.Data); err != nil {
				return err
			}
		}

		evts[i] = event
	}

	if err := s.publisher.PublishAsync(ctx, evts...); err != nil {
		return err
	}
	return nil
}

func (s *streamer) watch(ctx context.Context, stream store.Store_EventStreamClient) {
	for {
		select {
		case <-ctx.Done():
			if err := stream.CloseSend(); err != nil {
				log.Printf("error closing stream: %v", err)
			}
			return
		default:
			msg, err := stream.Recv()
			if err == io.EOF {
				// we've reached the end of the stream
				break
			}
			if err != nil {
				log.Fatalf("error while reading stream: %v", err)
			}
			if err := s.handle(ctx, msg); err != nil {
				log.Fatalf("error while handling stream: %v", err)
			}
		}
	}
}

func (s *streamer) Run(ctx context.Context) error {
	req := &store.EventStreamRequest{
		ServiceName: s.serviceName,
		EventTypes:  s.eventTypes,
	}

	stream, err := s.storeClient.EventStream(ctx, req)
	if err != nil {
		return err
	}

	// run it!.
	inner, cancel := context.WithCancel(ctx)
	go s.watch(inner, stream)
	s.cancel = cancel
	return nil
}

func (s *streamer) Close(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

func NewStreamer(storeClient store.StoreClient, cfg es.Config) Streamer {
	handlers := cfg.GetEventHandlers()
	serviceName := cfg.GetServiceName()

	publisher := es.NewPublisher(handlers)
	eventTypes := []string{}
	eventMap := map[string]reflect.Type{}

	for t := range handlers {
		eventTypes = append(eventTypes, t.String())
		eventMap[t.String()] = t
	}

	return &streamer{
		storeClient: storeClient,
		publisher:   publisher,
		eventTypes:  eventTypes,
		eventMap:    eventMap,
		serviceName: serviceName,
	}
}
