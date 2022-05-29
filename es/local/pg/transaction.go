package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"eventstore/es"

	"github.com/uptrace/bun"
)

type tx struct {
	*postgresData
	inner bun.Tx
}

func (t *tx) LoadSnapshot(ctx context.Context, serviceName string, aggregateName string, namespace string, id string, out es.SourcedAggregate) error {
	return nil
}
func (t *tx) GetEventDatas(ctx context.Context, serviceName string, aggregateName string, namespace string, id string, from int) ([]json.RawMessage, error) {
	// Select all users.
	var datas []json.RawMessage
	if err := t.inner.NewSelect().
		TableExpr("events").
		ColumnExpr("data").
		Where("service_name = ?", serviceName).
		Where("namespace = ?", namespace).
		Where("aggregate_type = ?", aggregateName).
		Where("aggregate_id = ?", id).
		Where("version > ?", from).
		Order("version").
		Scan(ctx, &datas); err != nil {
		if err != nil && sql.ErrNoRows != err {
			return nil, err
		}
	}
	return datas, nil
}
func (t *tx) SaveEvents(ctx context.Context, events []es.Event) error {
	if len(events) == 0 {
		return nil // nothing to save
	}

	dbEvents := make([]dbEvent, len(events))
	for i, evt := range events {
		dbEvents[i] = dbEvent{
			ServiceName:   evt.ServiceName,
			AggregateType: evt.AggregateType,
			Namespace:     evt.Namespace,
			AggregateId:   evt.AggregateId,
			Type:          evt.Type,
			Version:       evt.Version,
			Timestamp:     evt.Timestamp,
			Data:          evt.Data,
			Metadata:      evt.Metadata,
		}
	}

	// save em
	if _, err := t.inner.
		NewInsert().
		Model(&dbEvents).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}
func (t *tx) Commit(ctx context.Context) error {
	return t.inner.Commit()
}

func newTx(ctx context.Context, data *postgresData) (*tx, error) {
	inner, err := data.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &tx{
		postgresData: data,
		inner:        inner,
	}, nil
}
