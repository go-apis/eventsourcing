package local

import (
	"context"
	"database/sql"
	"encoding/json"
	"eventstore/es"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

type tx struct {
	*postgresData
	inner bun.Tx
}

func (t *tx) loadSnapshot(ctx context.Context, serviceName string, namespace string, id string, typeName string, out interface{}) (int, error) {
	return 0, nil
}
func (t *tx) loadEvents(ctx context.Context, serviceName string, namespace string, id string, typeName string, from int) ([]dbEvent, error) {
	// Select all users.
	var evts []dbEvent
	if err := t.inner.NewSelect().
		Model(&evts).
		Where("service_name = ?", serviceName).
		Where("namespace = ?", namespace).
		Where("aggregate_id = ?", id).
		Where("aggregate_type = ?", typeName).
		Where("version > ?", from).
		Order("version").
		Scan(ctx); err != nil {
		if err != nil && sql.ErrNoRows != err {
			return nil, err
		}
	}
	return evts, nil
}

func (t *tx) loadSourced(ctx context.Context, serviceName string, id string, typeName string, out es.SourcedAggregate) error {
	namespace := es.NamespaceFromContext(ctx)

	// load from snapshot
	version, err := t.loadSnapshot(ctx, serviceName, namespace, id, typeName, out)
	if err != nil {
		return err
	}

	// get the events
	events, err := t.loadEvents(ctx, serviceName, namespace, id, typeName, version)
	if err != nil {
		return err
	}

	for _, evt := range events {
		if err := json.Unmarshal(evt.Data, out); err != nil {
			return err
		}
		out.IncrementVersion()
	}

	return nil
}
func (t *tx) saveSourced(ctx context.Context, serviceName string, id string, typeName string, out es.SourcedAggregate) ([]es.Event, error) {
	datas := out.GetEvents()
	if len(datas) == 0 {
		return nil, nil // nothing to save
	}

	namespace := es.NamespaceFromContext(ctx)
	version := out.GetVersion()

	// get the events
	evts := make([]es.Event, len(datas))
	dbEvts := make([]dbEvent, len(datas))
	for i, data := range datas {
		buf, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		name := fmt.Sprintf("%T", data)
		metadata := es.MetadataFromContext(ctx)
		v := version + i + 1
		ts := time.Now()

		evts[i] = es.Event{
			ServiceName:   serviceName,
			Namespace:     namespace,
			AggregateId:   id,
			AggregateType: typeName,
			Type:          name,
			Version:       v,
			Timestamp:     ts,
			Data:          data,
			Metadata:      metadata,
		}
		dbEvts[i] = dbEvent{
			ServiceName:   serviceName,
			Namespace:     namespace,
			AggregateId:   id,
			AggregateType: typeName,
			Type:          name,
			Version:       v,
			Timestamp:     ts,
			Data:          buf,
			Metadata:      metadata,
		}
	}

	// save em
	if _, err := t.inner.
		NewInsert().
		Model(&dbEvts).
		Exec(ctx); err != nil {
		return nil, err
	}
	return evts, nil
}
func (t *tx) Load(ctx context.Context, id string, typeName string, out interface{}) error {
	switch impl := out.(type) {
	case es.SourcedAggregate:
		return t.loadSourced(ctx, id, typeName, impl)
	default:
		return fmt.Errorf("Invalid aggregate type")
	}
}
func (t *tx) Save(ctx context.Context, id string, typeName string, out interface{}) ([]es.Event, error) {
	switch impl := out.(type) {
	case es.SourcedAggregate:
		return t.saveSourced(ctx, id, typeName, impl)
	default:
		return nil, fmt.Errorf("Invalid aggregate type")
	}
}
func (t *tx) GetEvents(ctx context.Context) ([]es.Event, error) {
	// Select all users.
	var evts []es.Event
	if err := t.inner.NewSelect().
		Model(&evts).
		Order("timestamp desc").
		Scan(ctx); err != nil {
		if err != nil && sql.ErrNoRows != err {
			return nil, err
		}
	}
	return evts, nil
}
func (t *tx) Commit() error {
	return t.inner.Commit()
}

func newPostgresTx(ctx context.Context, data *postgresData) (*tx, error) {
	inner, err := data.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &tx{
		postgresData: data,
		inner:        inner,
	}, nil
}
