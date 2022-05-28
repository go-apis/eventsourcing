package es

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type SourcedStore interface {
	Load(ctx context.Context, id string) (SourcedAggregate, error)
	Save(ctx context.Context, id string, val SourcedAggregate) error
}

type sourceStore struct {
	e       *eventStore
	name    string
	factory SourcedAggregateFactory
}

func (s *sourceStore) Load(ctx context.Context, id string) (SourcedAggregate, error) {
	agg, err := s.factory()
	if err != nil {
		return nil, err
	}

	namespace := NamespaceFromContext(ctx)

	if err := s.e.Data.LoadSnapshot(ctx, s.e.serviceName, s.name, namespace, id, agg); err != nil {
		return nil, err
	}

	events, err := s.e.Data.GetEvents(ctx, s.e.serviceName, s.name, namespace, id, agg.GetVersion())
	if err != nil {
		return nil, err
	}

	for _, evt := range events {
		if err := json.Unmarshal(evt.Data, agg); err != nil {
			return nil, err
		}
		agg.IncrementVersion()
	}

	return agg, nil
}
func (s *sourceStore) Save(ctx context.Context, id string, val SourcedAggregate) error {
	namespace := NamespaceFromContext(ctx)

	// get the events
	evts := make([]es.Event, len(datas))
	for i, data := range datas {
		buf, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		name := fmt.Sprintf("%T", data)
		metadata := es.MetadataFromContext(ctx)
		v := version + i + 1
		ts := time.Now()

	}

	evts, err := s.e.Data.SaveSourced(ctx, s.e.serviceName, s.name, namespace, id, val)
	if err != nil {
		return err
	}

	if err := s.e.dispatcher.PublishAsync(ctx, evts...); err != nil {
		return err
	}

	return nil
}

func NewSourcedStore(e *eventStore, name string, factory SourcedAggregateFactory) SourcedStore {
	return &sourceStore{
		e:       e,
		name:    name,
		factory: factory,
	}
}
