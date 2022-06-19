package es

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type SourcedStore interface {
	Load(ctx context.Context, id string, out SourcedAggregate) error
	Save(ctx context.Context, id string, val SourcedAggregate) error
}

type sourceStore struct {
	tx            Tx
	publisher     Publisher
	serviceName   string
	aggregateName string
}

func (s *sourceStore) Load(ctx context.Context, id string, out SourcedAggregate) error {
	namespace := NamespaceFromContext(ctx)

	if err := s.tx.LoadSnapshot(ctx, s.serviceName, s.aggregateName, namespace, id, out); err != nil {
		return err
	}

	datas, err := s.tx.GetEventDatas(ctx, s.serviceName, s.aggregateName, namespace, id, out.GetVersion())
	if err != nil {
		return err
	}

	for _, data := range datas {
		if err := json.Unmarshal(data, out); err != nil {
			return err
		}
		out.IncrementVersion()
	}

	return nil
}
func (s *sourceStore) Save(ctx context.Context, id string, val SourcedAggregate) error {
	datas := val.GetEvents()
	version := val.GetVersion()
	namespace := NamespaceFromContext(ctx)

	if len(datas) == 0 {
		return nil
	}

	// get the events
	evts := make([]Event, len(datas))
	for i, data := range datas {
		name := fmt.Sprintf("%T", data)
		metadata := MetadataFromContext(ctx)
		v := version + i + 1
		ts := time.Now()

		evts[i] = Event{
			ServiceName:   s.serviceName,
			Namespace:     namespace,
			AggregateId:   id,
			AggregateType: s.aggregateName,
			Type:          name,
			Version:       v,
			Timestamp:     ts,
			Data:          data,
			Metadata:      metadata,
		}
	}

	if err := s.tx.SaveEvents(ctx, evts); err != nil {
		return err
	}

	if err := s.publisher.PublishAsync(ctx, evts...); err != nil {
		return err
	}

	return nil
}

func newSourcedStore(tx Tx, publisher Publisher, serviceName string, aggregateName string) SourcedStore {
	return &sourceStore{
		tx:            tx,
		publisher:     publisher,
		serviceName:   serviceName,
		aggregateName: aggregateName,
	}
}
