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
)

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

type DataStore interface {
	Load(ctx context.Context, id uuid.UUID, opts ...DataLoadOption) (Entity, error)
	Save(ctx context.Context, aggregate Entity) ([]*Event, error)
	Delete(ctx context.Context, aggregate Entity) error
	Truncate(ctx context.Context) error
}

type dataStore struct {
	data         Data
	entityConfig *EntityConfig
}

func (s *dataStore) applyEvents(ctx context.Context, aggregate AggregateSourced, events []*Event) error {
	for _, evt := range events {
		aggregate.IncrementVersion()

		t := reflect.TypeOf(evt.Data)
		h, ok := s.entityConfig.Handles[t]
		if ok {
			if err := h.Handle(aggregate, ctx, evt); err != nil {
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

func (s *dataStore) loadSourced(ctx context.Context, aggregate AggregateSourced, forced bool) (Entity, error) {
	namespace := NamespaceFromContext(ctx)
	id := aggregate.GetId()

	// load up the aggregate
	if s.entityConfig.SnapshotEnabled && s.entityConfig.SnapshotEvery >= 0 && !forced {
		snapshotSearch := SnapshotSearch{
			Namespace:     namespace,
			AggregateId:   id,
			AggregateType: s.entityConfig.Name,
			Revision:      s.entityConfig.SnapshotRevision,
		}
		if err := s.data.LoadSnapshot(ctx, snapshotSearch, aggregate); err != nil {
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
	originalEvents, err := s.data.GetEvents(ctx, s.entityConfig.Mapper, eventSearch)
	if err != nil {
		return nil, err
	}
	if err := s.applyEvents(ctx, aggregate, originalEvents); err != nil {
		return nil, err
	}
	return aggregate, nil
}
func (s *dataStore) loadEntity(ctx context.Context, entity Entity) (Entity, error) {
	namespace := NamespaceFromContext(ctx)

	id := entity.GetId()
	if err := s.data.Get(ctx, s.entityConfig.Name, namespace, id, entity); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	return entity, nil
}
func (s *dataStore) saveSourced(ctx context.Context, aggregate AggregateSourced) ([]*Event, error) {
	namespace := NamespaceFromContext(ctx)
	version := aggregate.GetVersion()
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
		v := version + i + 1
		metadata := MetadataFromContext(ctx)

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
	if err := s.applyEvents(ctx, aggregate, events); err != nil {
		return nil, err
	}

	if err := s.data.SaveEvents(ctx, events); err != nil {
		return nil, err
	}

	// save the snapshot!
	diff := aggregate.GetVersion() - version
	if diff < 0 {
		return nil, fmt.Errorf("version diff is less than 0")
	}

	if s.entityConfig.SnapshotEnabled && s.entityConfig.SnapshotEvery >= 0 && diff >= s.entityConfig.SnapshotEvery {
		snapshot := &Snapshot{
			Namespace:     namespace,
			AggregateId:   id,
			AggregateType: s.entityConfig.Name,
			Revision:      s.entityConfig.SnapshotRevision,
			Aggregate:     aggregate,
		}
		if err := s.data.SaveSnapshot(ctx, snapshot); err != nil {
			return nil, err
		}
	}

	if s.entityConfig.Project {
		if err := s.data.SaveEntity(ctx, s.entityConfig.Name, aggregate); err != nil {
			return nil, err
		}
	}

	return events, nil
}
func (s *dataStore) saveAggregateHolder(ctx context.Context, aggregate AggregateHolder) ([]*Event, error) {
	namespace := NamespaceFromContext(ctx)
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
		metadata := MetadataFromContext(ctx)

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

	if err := s.data.SaveEntity(ctx, s.entityConfig.Name, aggregate); err != nil {
		return nil, err
	}

	return events, nil
}
func (s *dataStore) saveEntity(ctx context.Context, aggregate Entity) ([]*Event, error) {
	return nil, s.data.SaveEntity(ctx, s.entityConfig.Name, aggregate)
}
func (s *dataStore) deleteEntity(ctx context.Context, aggregate Entity) error {
	return s.data.DeleteEntity(ctx, s.entityConfig.Name, aggregate)
}

func (s *dataStore) Load(ctx context.Context, id uuid.UUID, opts ...DataLoadOption) (Entity, error) {
	options := &DataLoadOptions{}
	for _, o := range opts {
		o(options)
	}
	namespace := NamespaceFromContext(ctx)

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
		return s.loadSourced(ctx, agg, options.Force)
	default:
		return s.loadEntity(ctx, agg)
	}
}
func (s *dataStore) Save(ctx context.Context, entity Entity) ([]*Event, error) {
	switch agg := entity.(type) {
	case AggregateSourced:
		return s.saveSourced(ctx, agg)
	case AggregateHolder:
		return s.saveAggregateHolder(ctx, agg)
	default:
		return s.saveEntity(ctx, agg)
	}
}
func (s *dataStore) Delete(ctx context.Context, entity Entity) error {
	switch agg := entity.(type) {
	case AggregateSourced:
		return fmt.Errorf("cannot delete an aggregate sourced entity")
	default:
		return s.deleteEntity(ctx, agg)
	}
}

func (s *dataStore) Truncate(ctx context.Context) error {
	return s.data.Truncate(ctx, s.entityConfig.Name)
}

// NewDataStore for creating stores
func NewDataStore(data Data, entityConfig *EntityConfig) DataStore {
	s := &dataStore{
		data:         data,
		entityConfig: entityConfig,
	}
	return s
}
