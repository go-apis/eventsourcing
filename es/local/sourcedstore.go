package local

import (
	"context"
	"encoding/json"
	"eventstore/es"
	"fmt"
	"time"
)

type sourceStore struct {
	data          es.Data
	dispatcher    es.Dispatcher
	serviceName   string
	aggregateName string
}

func (s *sourceStore) Load(ctx context.Context, id string, out es.SourcedAggregate) error {
	namespace := es.NamespaceFromContext(ctx)

	if err := s.data.LoadSnapshot(ctx, s.serviceName, s.aggregateName, namespace, id, out); err != nil {
		return err
	}

	datas, err := s.data.GetEventDatas(ctx, s.serviceName, s.aggregateName, namespace, id, out.GetVersion())
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
func (s *sourceStore) Save(ctx context.Context, id string, val es.SourcedAggregate) error {
	datas := val.GetEvents()
	version := val.GetVersion()
	namespace := es.NamespaceFromContext(ctx)

	if len(datas) == 0 {
		return nil
	}

	// get the events
	evts := make([]es.Event, len(datas))
	for i, data := range datas {
		name := fmt.Sprintf("%T", data)
		metadata := es.MetadataFromContext(ctx)
		v := version + i + 1
		ts := time.Now()

		evts[i] = es.Event{
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

	if err := s.data.SaveEvents(ctx, evts); err != nil {
		return err
	}

	if err := s.dispatcher.PublishAsync(ctx, evts...); err != nil {
		return err
	}

	return nil
}

func newSourcedStore(data es.Data, serviceName string, dispatcher es.Dispatcher, aggregateName string) es.SourcedStore {
	return &sourceStore{
		data:          data,
		serviceName:   serviceName,
		dispatcher:    dispatcher,
		aggregateName: aggregateName,
	}
}
