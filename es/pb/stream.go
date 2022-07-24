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

func (s *streamer) handle(ctx context.Context, evt *store.Event) error {
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

	if err := s.publisher.PublishAsync(ctx, event); err != nil {
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
				// todo how do we reconnect
				// we've reached the end of the stream
				break
			}
			if err != nil {
				log.Fatalf("error while reading stream: %v", err)
				continue
			}
			if err := s.handle(ctx, msg.Event); err != nil {
				log.Fatalf("error while handling stream: %v", err)

				// send nack
				nack := &store.EventStreamRequest{
					Msg: &store.EventStreamRequest_NackMsg{
						NackMsg: &store.EventId{
							ServiceName:   msg.Event.ServiceName,
							Namespace:     msg.Event.Namespace,
							AggregateType: msg.Event.AggregateType,
							AggregateId:   msg.Event.AggregateId,
							Type:          msg.Event.Type,
							Version:       msg.Event.Version,
						},
					},
				}
				if err := stream.Send(nack); err != nil {
					log.Fatalf("error while sending nack: %v", err)
				}
				continue
			}

			// send ack
			ack := &store.EventStreamRequest{
				Msg: &store.EventStreamRequest_AckMsg{
					AckMsg: &store.EventId{
						ServiceName:   msg.Event.ServiceName,
						Namespace:     msg.Event.Namespace,
						AggregateType: msg.Event.AggregateType,
						AggregateId:   msg.Event.AggregateId,
						Type:          msg.Event.Type,
						Version:       msg.Event.Version,
					},
				},
			}
			if err := stream.Send(ack); err != nil {
				log.Fatalf("error while sending ack: %v", err)
			}
		}
	}
}

func (s *streamer) Run(ctx context.Context) error {
	stream, err := s.storeClient.EventStream(ctx)
	if err != nil {
		return err
	}

	// run it!.
	inner, cancel := context.WithCancel(ctx)
	go s.watch(inner, stream)
	s.cancel = cancel

	// send the first event.
	req := &store.EventStreamRequest{
		Msg: &store.EventStreamRequest_InitMsg{
			InitMsg: &store.EventStreamRequest_Init{
				ServiceName: s.serviceName,
				EventTypes:  s.eventTypes,
			},
		},
	}

	if err := stream.Send(req); err != nil {
		return err
	}
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
