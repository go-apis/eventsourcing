package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/filters"
	"github.com/contextcloud/eventstore/pkg/pgdb"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type data struct {
	serviceName string
	db          *gorm.DB
	tx          *gorm.DB

	isCommitted bool
}

func (d *data) getDb() *gorm.DB {
	if d.tx != nil && !d.isCommitted {
		return d.tx
	}
	return d.db
}

func (d *data) inTransaction() bool {
	return d.tx != nil && !d.isCommitted
}

func (d *data) Begin(ctx context.Context) (es.Tx, error) {
	_, span := otel.Tracer("local").Start(ctx, "Begin")
	defer span.End()

	if d.isCommitted {
		return nil, fmt.Errorf("cannot begin transaction after commit")
	}

	if d.tx == nil {
		tx := d.db.Begin()
		if tx.Error != nil {
			return nil, tx.Error
		}
		d.tx = tx
	}

	return newTransaction(d), nil
}

func (d *data) LoadSnapshot(ctx context.Context, search es.SnapshotSearch, out es.AggregateSourced) error {
	pctx, span := otel.Tracer("local").Start(ctx, "LoadSnapshot")
	defer span.End()

	var snapshot pgdb.Snapshot
	r := d.getDb().
		WithContext(pctx).
		Model(&pgdb.Snapshot{}).
		Where("service_name = ?", d.serviceName).
		Where("namespace = ?", search.Namespace).
		Where("aggregate_type = ?", search.AggregateType).
		Where("aggregate_id = ?", search.AggregateId).
		Where("revision = ?", search.Revision).
		First(&snapshot)
	if r.Error == gorm.ErrRecordNotFound {
		return nil
	}
	if r.Error != nil {
		return r.Error
	}

	if err := json.Unmarshal(snapshot.Aggregate, out); err != nil {
		return err
	}

	return r.Error
}
func (d *data) SaveSnapshot(ctx context.Context, snapshot *es.Snapshot) error {
	pctx, span := otel.Tracer("local").Start(ctx, "SaveSnapshot")
	defer span.End()

	if !d.inTransaction() {
		return fmt.Errorf("must be in transaction")
	}

	if snapshot == nil {
		return nil // nothing to save
	}

	raw, err := json.Marshal(snapshot.Aggregate)
	if err != nil {
		return err
	}

	obj := &pgdb.Snapshot{
		ServiceName:   d.serviceName,
		Namespace:     snapshot.Namespace,
		AggregateId:   snapshot.AggregateId,
		AggregateType: snapshot.AggregateType,
		Revision:      snapshot.Revision,
		Aggregate:     raw,
	}

	out := d.getDb().
		WithContext(pctx).
		Clauses(clause.OnConflict{
			UpdateAll: true,
		}).
		Create(obj)
	return out.Error
}

func (d *data) loadData(mappers es.EventDataMapper, evt *pgdb.Event) (interface{}, error) {
	mapper, ok := mappers[evt.Type]
	if !ok {
		return evt.Data, nil
	}

	data, err := mapper()
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(evt.Data, data); err != nil {
		return nil, err
	}

	return data, nil
}

func (d *data) GetEvents(ctx context.Context, mappers es.EventDataMapper, search es.EventSearch) ([]*es.Event, error) {
	pctx, span := otel.Tracer("local").Start(ctx, "GetEvents")
	defer span.End()

	g := d.getDb().
		WithContext(pctx)

	rows, err := g.
		Model(&pgdb.Event{}).
		Where("service_name = ?", d.serviceName).
		Where("namespace = ?", search.Namespace).
		Where("aggregate_type = ?", search.AggregateType).
		Where("aggregate_id = ?", search.AggregateId).
		Where("version > ?", search.FromVersion).
		Order("version").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*es.Event
	for rows.Next() {
		var evt pgdb.Event
		// ScanRows is a method of `gorm.DB`, it can be used to scan a row into a struct
		if err := g.ScanRows(rows, &evt); err != nil {
			return nil, err
		}

		// do something
		data, err := d.loadData(mappers, &evt)
		if err != nil {
			return nil, err
		}

		events = append(events, &es.Event{
			// ServiceName:   d.serviceName,
			Namespace:     evt.Namespace,
			AggregateId:   evt.AggregateId,
			AggregateType: evt.AggregateType,
			Type:          evt.Type,
			Version:       evt.Version,
			Timestamp:     evt.Timestamp,
			Data:          data,
			Metadata:      evt.Metadata,
		})
	}

	return events, nil
}
func (d *data) SaveEvents(ctx context.Context, events []*es.Event) error {
	pctx, span := otel.Tracer("local").Start(ctx, "SaveEvents")
	defer span.End()

	if !d.inTransaction() {
		return fmt.Errorf("must be in transaction")
	}

	if len(events) == 0 {
		return nil // nothing to save
	}

	evts := make([]*pgdb.Event, len(events))
	for i, evt := range events {
		raw, err := json.Marshal(evt.Data)
		if err != nil {
			return err
		}

		evts[i] = &pgdb.Event{
			ServiceName:   d.serviceName,
			Namespace:     evt.Namespace,
			AggregateId:   evt.AggregateId,
			AggregateType: evt.AggregateType,
			Type:          evt.Type,
			Version:       evt.Version,
			Timestamp:     evt.Timestamp,
			Data:          raw,
			Metadata:      evt.Metadata,
		}
	}

	out := d.getDb().
		WithContext(pctx).
		Create(&evts)
	return out.Error
}
func (d *data) SaveEntity(ctx context.Context, aggregateName string, raw es.Entity) error {
	pctx, span := otel.Tracer("local").Start(ctx, "SaveEntity")
	defer span.End()

	if !d.inTransaction() {
		return fmt.Errorf("must be in transaction")
	}

	table := pgdb.TableName(d.serviceName, aggregateName)
	out := d.getDb().
		WithContext(pctx).
		Table(table).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}, {Name: "namespace"}},
			UpdateAll: true,
		}).
		Create(raw)
	return out.Error
}
func (d *data) DeleteEntity(ctx context.Context, aggregateName string, raw es.Entity) error {
	pctx, span := otel.Tracer("local").Start(ctx, "DeleteEntity")
	defer span.End()

	if !d.inTransaction() {
		return fmt.Errorf("must be in transaction")
	}

	table := pgdb.TableName(d.serviceName, aggregateName)
	out := d.getDb().
		WithContext(pctx).
		Table(table).
		Delete(raw, "namespace = ?", raw.GetNamespace())
	return out.Error
}

func (d *data) Get(ctx context.Context, aggregateName string, namespace string, id uuid.UUID, out interface{}) error {
	pctx, span := otel.Tracer("local").Start(ctx, "Load")
	defer span.End()

	table := pgdb.TableName(d.serviceName, aggregateName)

	q := d.getDb().
		WithContext(pctx).
		Table(table).
		Where("id = ?", id)

	if namespace != "" {
		q = q.Where("namespace = ?", namespace)
	}

	r := q.First(out)
	if r.Error == gorm.ErrRecordNotFound {
		return sql.ErrNoRows
	}
	return r.Error
}
func (d *data) Find(ctx context.Context, aggregateName string, namespace string, filter filters.Filter, out interface{}) error {
	pctx, span := otel.Tracer("local").Start(ctx, "Find")
	defer span.End()

	table := pgdb.TableName(d.serviceName, aggregateName)
	q := d.getDb().
		WithContext(pctx).
		Table(table)

	q = where(q, filter.Where)

	if namespace != "" {
		q = q.Where("namespace = ?", namespace)
	}

	if filter.Distinct != nil {
		q = q.Distinct(filter.Distinct...)
	}

	if filter.Limit != nil {
		q = q.Limit(*filter.Limit)
	}
	if filter.Offset != nil {
		q = q.Offset(*filter.Offset)
	}

	for _, order := range filter.Order {
		q = q.Order(clause.OrderByColumn{
			Column: clause.Column{
				Name: order.Column,
			},
			Desc: order.Desc,
		})
	}

	r := q.
		Find(out)
	return r.Error
}
func (d *data) Count(ctx context.Context, aggregateName string, namespace string, filter filters.Filter) (int, error) {
	pctx, span := otel.Tracer("local").Start(ctx, "Count")
	defer span.End()

	var totalRows int64

	table := pgdb.TableName(d.serviceName, aggregateName)
	q := d.getDb().
		WithContext(pctx).
		Table(table)

	q = where(q, filter.Where)

	if namespace != "" {
		q = q.Where("namespace = ?", namespace)
	}

	if filter.Distinct != nil {
		q = q.Distinct(filter.Distinct...)
	}

	r := q.Count(&totalRows)
	return int(totalRows), r.Error
}

func newData(serviceName string, db *gorm.DB) es.Data {
	d := &data{
		serviceName: serviceName,
		db:          db,
	}
	return d
}
