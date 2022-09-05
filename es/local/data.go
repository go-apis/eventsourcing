package local

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/filters"
	"github.com/contextcloud/eventstore/pkg/db"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type data struct {
	db *gorm.DB
	tx *gorm.DB

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

func (t *data) LoadSnapshot(ctx context.Context, serviceName string, aggregateName string, namespace string, revision string, id uuid.UUID, out es.AggregateSourced) error {
	return nil
}
func (d *data) SaveSnapshot(ctx context.Context, serviceName string, aggregateName string, namespace string, revision string, id uuid.UUID, out es.AggregateSourced) error {
	return nil
}

func (d *data) loadData(mappers es.EventDataMapper, evt *db.Event) (interface{}, error) {
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
	pctx, span := otel.Tracer("local").Start(ctx, "GetEventDatas")
	defer span.End()

	g := d.getDb().
		WithContext(pctx)

	rows, err := g.
		Model(&db.Event{}).
		Where("service_name = ?", search.ServiceName).
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
		var evt db.Event
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
			ServiceName:   evt.ServiceName,
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
	pctx, span := otel.Tracer("local").Start(ctx, "SaveEventDatas")
	defer span.End()

	if !d.inTransaction() {
		return fmt.Errorf("must be in transaction")
	}

	if len(events) == 0 {
		return nil // nothing to save
	}

	evts := make([]*db.Event, len(events))
	for i, d := range events {
		raw, err := json.Marshal(d.Data)
		if err != nil {
			return err
		}

		evts[i] = &db.Event{
			ServiceName:   d.ServiceName,
			Namespace:     d.Namespace,
			AggregateId:   d.AggregateId,
			AggregateType: d.AggregateType,
			Type:          d.Type,
			Version:       d.Version,
			Timestamp:     d.Timestamp,
			Data:          raw,
			Metadata:      d.Metadata,
		}
	}

	out := d.getDb().
		WithContext(pctx).
		Create(&evts)
	return out.Error
}
func (t *data) SaveEntity(ctx context.Context, serviceName string, aggregateName string, raw es.Entity) error {
	pctx, span := otel.Tracer("local").Start(ctx, "SaveEntity")
	defer span.End()

	if !t.inTransaction() {
		return fmt.Errorf("must be in transaction")
	}

	table := db.TableName(serviceName, aggregateName)
	out := t.getDb().
		WithContext(pctx).
		Table(table).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}, {Name: "namespace"}},
			UpdateAll: true,
		}).
		Create(raw)
	return out.Error
}

func (t *data) Load(ctx context.Context, serviceName string, aggregateName string, namespace string, id uuid.UUID, out interface{}) error {
	pctx, span := otel.Tracer("local").Start(ctx, "Load")
	defer span.End()

	table := db.TableName(serviceName, aggregateName)

	r := t.getDb().
		WithContext(pctx).
		Table(table).
		Where("id = ?", id).
		Where("namespace = ?", namespace).
		First(out)
	return r.Error
}
func (t *data) Find(ctx context.Context, serviceName string, aggregateName string, namespace string, filter filters.Filter, out interface{}) error {
	pctx, span := otel.Tracer("local").Start(ctx, "Find")
	defer span.End()

	table := db.TableName(serviceName, aggregateName)
	q := t.getDb().
		WithContext(pctx).
		Table(table).
		Where("namespace = ?", namespace)

	q = where(q, filter.Where)

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
func (t *data) Count(ctx context.Context, serviceName string, aggregateName string, namespace string, filter filters.Filter) (int, error) {
	pctx, span := otel.Tracer("local").Start(ctx, "Count")
	defer span.End()

	var totalRows int64

	table := db.TableName(serviceName, aggregateName)
	q := t.getDb().
		WithContext(pctx).
		Table(table).
		Where("namespace = ?", namespace)

	q = where(q, filter.Where)
	r := q.Count(&totalRows)
	return int(totalRows), r.Error
}

func newData(db *gorm.DB) es.Data {
	d := &data{
		db: db,
	}
	return d
}
