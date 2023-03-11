package es

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

type IsApplyEvent interface {
	ApplyEvent(ctx context.Context, evt *Event) error
}

func toJson(data interface{}) (json.RawMessage, error) {
	switch t := data.(type) {
	case []byte:
		return t, nil
	case json.RawMessage:
		return t, nil
	default:
		return json.Marshal(t)
	}
}

func applyEvents(ctx context.Context, aggregate AggregateSourced, events []*Event) error {
	for _, evt := range events {
		aggregate.IncrementVersion()

		a, ok := aggregate.(IsApplyEvent)
		if ok {
			if err := a.ApplyEvent(ctx, evt); err != nil {
				return err
			}
			continue
		}

		raw, err := toJson(evt.Data)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(raw, aggregate); err != nil {
			return err
		}
	}

	return nil
}

type DataStore interface {
	Load(ctx context.Context, id uuid.UUID, opts ...DataLoadOption) (Entity, error)
	Save(ctx context.Context, aggregate Entity) ([]*Event, error)
}

type dataStore struct {
	data         Data
	entityConfig *EntityConfig
}

func (s *dataStore) loadSourced(ctx context.Context, aggregate AggregateSourced, forced bool) (Entity, error) {
	pctx, pspan := otel.Tracer("DataStore").Start(ctx, "LoadSourced")
	defer pspan.End()

	namespace := NamespaceFromContext(pctx)
	id := aggregate.GetId()

	// load up the aggregate
	if s.entityConfig.SnapshotEvery >= 0 && !forced {
		snapshotSearch := SnapshotSearch{
			Namespace:     namespace,
			AggregateId:   id,
			AggregateType: s.entityConfig.Name,
			Revision:      s.entityConfig.SnapshotRevision,
		}
		if err := s.data.LoadSnapshot(pctx, snapshotSearch, aggregate); err != nil {
			return nil, err
		}
	}

	// load up the events from the DB.
	eventSearch := EventSearch{
		Namespace:     namespace,
		AggregateId:   id,
		AggregateType: s.entityConfig.Name,
		FromVersion:   aggregate.GetVersion(),
	}
	originalEvents, err := s.data.GetEvents(pctx, s.entityConfig.Mapper, eventSearch)
	if err != nil {
		return nil, err
	}
	if err := applyEvents(pctx, aggregate, originalEvents); err != nil {
		return nil, err
	}
	return aggregate, nil
}
func (s *dataStore) loadEntity(ctx context.Context, entity Entity) (Entity, error) {
	pctx, pspan := otel.Tracer("DataStore").Start(ctx, "LoadEntity")
	defer pspan.End()

	namespace := NamespaceFromContext(pctx)

	id := entity.GetId()
	if err := s.data.Get(pctx, s.entityConfig.Name, namespace, id, entity); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	return entity, nil
}
func (s *dataStore) saveSourced(ctx context.Context, aggregate AggregateSourced) ([]*Event, error) {
	pctx, pspan := otel.Tracer("DataStore").Start(ctx, "SaveSourced")
	defer pspan.End()

	namespace := NamespaceFromContext(pctx)
	version := aggregate.GetVersion()
	id := aggregate.GetId()
	raw := aggregate.GetEvents()
	timestamp := time.Now()
	_, hasApplyEvent := aggregate.(IsApplyEvent)

	events := make([]*Event, len(raw))
	for i, data := range raw {
		t := reflect.TypeOf(data)
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		name := t.Name()
		v := version + i + 1
		metadata := MetadataFromContext(pctx)

		// validate if we have a way to build the event with our mapper.
		if hasApplyEvent {
			if _, ok := s.entityConfig.Mapper[name]; !ok {
				return nil, fmt.Errorf("no mapper function for event %s", name)
			}
		}

		events[i] = &Event{
			Namespace:     namespace,
			AggregateId:   id,
			AggregateType: s.entityConfig.Name,
			Type:          name,
			Version:       v,
			Data:          data,
			Timestamp:     timestamp,
			Metadata:      metadata,
		}
	}

	// Apply the events so we can save the aggregate
	if err := applyEvents(pctx, aggregate, events); err != nil {
		return nil, err
	}

	if err := s.data.SaveEvents(pctx, events); err != nil {
		return nil, err
	}

	// save the snapshot!
	diff := aggregate.GetVersion() - version
	if diff < 0 {
		return nil, fmt.Errorf("version diff is less than 0")
	}

	if s.entityConfig.SnapshotEvery >= 0 && diff >= s.entityConfig.SnapshotEvery {
		snapshot := &Snapshot{
			Namespace:     namespace,
			AggregateId:   id,
			AggregateType: s.entityConfig.Name,
			Revision:      s.entityConfig.SnapshotRevision,
			Aggregate:     aggregate,
		}
		if err := s.data.SaveSnapshot(pctx, snapshot); err != nil {
			return nil, err
		}
	}

	if s.entityConfig.Project {
		if err := s.data.SaveEntity(pctx, s.entityConfig.Name, aggregate); err != nil {
			return nil, err
		}
	}

	return events, nil
}
func (s *dataStore) saveAggregateHolder(ctx context.Context, aggregate AggregateHolder) ([]*Event, error) {
	pctx, pspan := otel.Tracer("DataStore").Start(ctx, "SaveAggregateHolder")
	defer pspan.End()

	namespace := NamespaceFromContext(pctx)
	id := aggregate.GetId()
	raw := aggregate.GetEvents()
	timestamp := time.Now()

	events := make([]*Event, len(raw))
	for i, data := range raw {
		t := reflect.TypeOf(data)
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		name := t.Name()
		metadata := MetadataFromContext(pctx)

		events[i] = &Event{
			Namespace:     namespace,
			AggregateId:   id,
			AggregateType: s.entityConfig.Name,
			Type:          name,
			Data:          data,
			Timestamp:     timestamp,
			Metadata:      metadata,
		}
	}

	if err := s.data.SaveEntity(pctx, s.entityConfig.Name, aggregate); err != nil {
		return nil, err
	}

	return events, nil
}
func (s *dataStore) saveEntity(ctx context.Context, aggregate Entity) ([]*Event, error) {
	pctx, pspan := otel.Tracer("DataStore").Start(ctx, "SaveEntity")
	defer pspan.End()

	return nil, s.data.SaveEntity(pctx, s.entityConfig.Name, aggregate)
}

func (s *dataStore) Load(ctx context.Context, id uuid.UUID, opts ...DataLoadOption) (Entity, error) {
	pctx, pspan := otel.Tracer("DataStore").Start(ctx, "Load")
	defer pspan.End()

	options := &DataLoadOptions{}
	for _, o := range opts {
		o(options)
	}
	namespace := NamespaceFromContext(pctx)

	// make the aggregate
	entity, err := s.entityConfig.Factory()
	if err != nil {
		return nil, err
	}

	switch agg := entity.(type) {
	case SetId:
		agg.SetId(id, namespace)
	}

	switch agg := entity.(type) {
	case AggregateSourced:
		return s.loadSourced(pctx, agg, options.Force)
	default:
		return s.loadEntity(pctx, agg)
	}
}
func (s *dataStore) Save(ctx context.Context, entity Entity) ([]*Event, error) {
	pctx, pspan := otel.Tracer("DataStore").Start(ctx, "Save")
	defer pspan.End()

	switch agg := entity.(type) {
	case AggregateSourced:
		return s.saveSourced(pctx, agg)
	case AggregateHolder:
		return s.saveAggregateHolder(pctx, agg)
	default:
		return s.saveEntity(pctx, agg)
	}
}

// NewDataStore for creating stores
func NewDataStore(data Data, entityConfig *EntityConfig) DataStore {
	s := &dataStore{
		data:         data,
		entityConfig: entityConfig,
	}
	return s
}
