package es

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/contextcloud/eventstore/es/filters"
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
	Load(ctx context.Context, name string, id uuid.UUID, opts ...DataLoadOption) (Entity, error)
	Save(ctx context.Context, name string, aggregate Entity) ([]*Event, error)
	Delete(ctx context.Context, name string, aggregate Entity) error
	Truncate(ctx context.Context, name string) error
}

type dataStore struct {
	service  string
	data     Data
	registry Registry
}

func (s *dataStore) applyEvents(ctx context.Context, entityConfig *EntityConfig, aggregate AggregateSourced, events []*Event) error {
	for _, evt := range events {
		aggregate.IncrementVersion()

		t := reflect.TypeOf(evt.Data)
		h, ok := entityConfig.Handles[t]
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
func (s *dataStore) loadSourced(ctx context.Context, entityConfig *EntityConfig, aggregate AggregateSourced, forced bool) (Entity, error) {
	namespace := GetNamespace(ctx)
	id := aggregate.GetId()

	// load up the aggregate
	if entityConfig.SnapshotEnabled && entityConfig.SnapshotEvery >= 0 && !forced {
		snapshotSearch := SnapshotSearch{
			Namespace:     namespace,
			AggregateId:   id,
			AggregateType: entityConfig.Name,
			Revision:      entityConfig.SnapshotRevision,
		}
		if err := s.data.LoadSnapshot(ctx, snapshotSearch, aggregate); err != nil {
			return nil, err
		}
	}

	eventFilter := filters.Filter{
		Where: []filters.WhereClause{
			{
				Column: "namespace",
				Op:     "eq",
				Args:   namespace,
			},
			{
				Column: "aggregate_id",
				Op:     "eq",
				Args:   id,
			},
			{
				Column: "aggregate_type",
				Op:     "eq",
				Args:   entityConfig.Name,
			},
			{
				Column: "version",
				Op:     "gt",
				Args:   aggregate.GetVersion(),
			},
		},
	}
	// load up the events from the DB.
	originalEvents, err := s.data.FindEvents(ctx, eventFilter)
	if err != nil {
		return nil, err
	}
	if err := s.applyEvents(ctx, entityConfig, aggregate, originalEvents); err != nil {
		return nil, err
	}
	return aggregate, nil
}
func (s *dataStore) loadEntity(ctx context.Context, entityConfig *EntityConfig, entity Entity) (Entity, error) {
	namespace := GetNamespace(ctx)

	id := entity.GetId()
	if err := s.data.Get(ctx, entityConfig.Name, namespace, id, entity); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	return entity, nil
}
func (s *dataStore) saveSourced(ctx context.Context, entityConfig *EntityConfig, aggregate AggregateSourced) ([]*Event, error) {
	namespace := GetNamespace(ctx)
	version := aggregate.GetVersion()
	id := aggregate.GetId()
	raw := aggregate.GetEvents()
	timestamp := GetTime(ctx)

	events := make([]*Event, len(raw))
	for i, data := range raw {
		t := reflect.TypeOf(data)
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		name := t.Name()
		v := version + i + 1
		metadata := GetMetadata(ctx)

		events[i] = &Event{
			Service:       s.service,
			Namespace:     namespace,
			AggregateId:   id,
			AggregateType: entityConfig.Name,
			Type:          name,
			Version:       v,
			Data:          data,
			Timestamp:     timestamp,
			Metadata:      metadata,
		}
	}

	// Apply the events so we can save the aggregate
	if err := s.applyEvents(ctx, entityConfig, aggregate, events); err != nil {
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

	if entityConfig.SnapshotEnabled && entityConfig.SnapshotEvery >= 0 && diff >= entityConfig.SnapshotEvery {
		snapshot := &Snapshot{
			Namespace:     namespace,
			AggregateId:   id,
			AggregateType: entityConfig.Name,
			Revision:      entityConfig.SnapshotRevision,
			Aggregate:     aggregate,
		}
		if err := s.data.SaveSnapshot(ctx, snapshot); err != nil {
			return nil, err
		}
	}

	if entityConfig.Project {
		if err := s.data.SaveEntity(ctx, entityConfig.Name, aggregate); err != nil {
			return nil, err
		}
	}

	return events, nil
}
func (s *dataStore) saveAggregateHolder(ctx context.Context, entityConfig *EntityConfig, aggregate AggregateHolder) ([]*Event, error) {
	namespace := GetNamespace(ctx)
	id := aggregate.GetId()
	raw := aggregate.GetEvents()
	timestamp := GetTime(ctx)

	events := make([]*Event, len(raw))
	for i, data := range raw {
		t := reflect.TypeOf(data)
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		name := t.Name()
		metadata := GetMetadata(ctx)

		events[i] = &Event{
			Service:       s.service,
			Namespace:     namespace,
			AggregateId:   id,
			AggregateType: entityConfig.Name,
			Type:          name,
			Data:          data,
			Timestamp:     timestamp,
			Metadata:      metadata,
		}
	}

	if err := s.data.SaveEntity(ctx, entityConfig.Name, aggregate); err != nil {
		return nil, err
	}

	return events, nil
}
func (s *dataStore) saveEntity(ctx context.Context, entityConfig *EntityConfig, aggregate Entity) ([]*Event, error) {
	return nil, s.data.SaveEntity(ctx, entityConfig.Name, aggregate)
}
func (s *dataStore) deleteEntity(ctx context.Context, entityConfig *EntityConfig, aggregate Entity) error {
	return s.data.DeleteEntity(ctx, entityConfig.Name, aggregate)
}

func (s *dataStore) Load(ctx context.Context, name string, id uuid.UUID, opts ...DataLoadOption) (Entity, error) {
	entityConfig, err := s.registry.GetEntityConfig(name)
	if err != nil {
		return nil, err
	}

	options := &DataLoadOptions{}
	for _, o := range opts {
		o(options)
	}
	namespace := GetNamespace(ctx)

	// make the aggregate
	entity, err := entityConfig.Factory()
	if err != nil {
		return nil, err
	}

	switch agg := entity.(type) {
	case SetId:
		agg.SetId(id, namespace)
	}

	switch agg := entity.(type) {
	case AggregateSourced:
		return s.loadSourced(ctx, entityConfig, agg, options.Force)
	default:
		return s.loadEntity(ctx, entityConfig, agg)
	}
}
func (s *dataStore) Save(ctx context.Context, name string, entity Entity) ([]*Event, error) {
	entityConfig, err := s.registry.GetEntityConfig(name)
	if err != nil {
		return nil, err
	}

	switch agg := entity.(type) {
	case AggregateSourced:
		return s.saveSourced(ctx, entityConfig, agg)
	case AggregateHolder:
		return s.saveAggregateHolder(ctx, entityConfig, agg)
	default:
		return s.saveEntity(ctx, entityConfig, agg)
	}
}
func (s *dataStore) Delete(ctx context.Context, name string, entity Entity) error {
	entityConfig, err := s.registry.GetEntityConfig(name)
	if err != nil {
		return err
	}

	switch agg := entity.(type) {
	case AggregateSourced:
		return fmt.Errorf("cannot delete an aggregate sourced entity")
	default:
		return s.deleteEntity(ctx, entityConfig, agg)
	}
}
func (s *dataStore) Truncate(ctx context.Context, name string) error {
	entityConfig, err := s.registry.GetEntityConfig(name)
	if err != nil {
		return err
	}
	return s.data.Truncate(ctx, entityConfig.Name)
}

// NewDataStore for creating stores
func NewDataStore(service string, data Data, reg Registry) DataStore {
	s := &dataStore{
		service:  service,
		data:     data,
		registry: reg,
	}
	return s
}
