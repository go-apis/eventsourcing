package es

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type SourcedStore interface {
	Load(ctx context.Context, id string, namespace string, out SourcedAggregate) error
	Save(ctx context.Context, id string, namespace string, val SourcedAggregate) error
}

type sourcedStore struct {
	publisher     Publisher
	serviceName   string
	aggregateName string
}

func (s *sourcedStore) Load(ctx context.Context, id string, namespace string, out SourcedAggregate) error {
	tx := TxFromContext(ctx)

	if err := tx.LoadSnapshot(ctx, s.serviceName, s.aggregateName, namespace, id, out); err != nil {
		return err
	}

	datas, err := tx.GetEventDatas(ctx, s.serviceName, s.aggregateName, namespace, id, out.GetVersion())
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
func (s *sourcedStore) Save(ctx context.Context, id string, namespace string, val SourcedAggregate) error {
	datas := val.GetEvents()
	version := val.GetVersion()
	tx := TxFromContext(ctx)

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

	if err := tx.SaveEvents(ctx, evts); err != nil {
		return err
	}

	// HACK
	for _, evt := range evts {
		raw, err := json.Marshal(evt.Data)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(raw, val); err != nil {
			return err
		}
		val.IncrementVersion()
	}

	// Save read view.
	entity := Entity{
		ServiceName:   s.serviceName,
		Namespace:     namespace,
		AggregateId:   id,
		AggregateType: s.aggregateName,
		Data:          val,
	}
	if err := tx.SaveEntity(ctx, entity); err != nil {
		return err
	}

	if err := s.publisher.PublishAsync(ctx, evts...); err != nil {
		return err
	}

	return nil
}

func NewSourcedStore(publisher Publisher, serviceName string, aggregateName string) SourcedStore {
	return &sourcedStore{
		publisher:     publisher,
		serviceName:   serviceName,
		aggregateName: aggregateName,
	}
}
