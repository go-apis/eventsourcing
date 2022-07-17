package local

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/filters"
	"github.com/contextcloud/eventstore/pkg/db"

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

func (t *data) LoadSnapshot(ctx context.Context, serviceName string, aggregateName string, namespace string, id string, out es.SourcedAggregate) error {
	return nil
}
func (t *data) GetEventDatas(ctx context.Context, serviceName string, aggregateName string, namespace string, id string, fromVersion int) ([]json.RawMessage, error) {
	var datas []json.RawMessage
	out := t.getDb().WithContext(ctx).
		Where("service_name = ?", serviceName).
		Where("namespace = ?", namespace).
		Where("aggregate_type = ?", aggregateName).
		Where("aggregate_id = ?", id).
		Where("version > ?", fromVersion).
		Order("version").
		Model(&db.Event{}).
		Pluck("data", &datas)
	return datas, out.Error
}
func (t *data) SaveEvents(ctx context.Context, events []es.Event) error {
	if !t.inTransaction() {
		return fmt.Errorf("must be in transaction")
	}

	if len(events) == 0 {
		return nil // nothing to save
	}

	data := make([]*db.Event, len(events))
	for i, evt := range events {
		data[i] = &db.Event{
			ServiceName:   evt.ServiceName,
			Namespace:     evt.Namespace,
			AggregateId:   evt.AggregateId,
			AggregateType: evt.AggregateType,
			Type:          evt.Type,
			Version:       evt.Version,
			Timestamp:     evt.Timestamp,
			Data:          evt.Data,
			Metadata:      evt.Metadata,
		}
	}

	out := t.getDb().WithContext(ctx).
		Create(&data)
	return out.Error
}
func (t *data) SaveEntity(ctx context.Context, raw es.Entity) error {
	if !t.inTransaction() {
		return fmt.Errorf("must be in transaction")
	}

	table := db.TableName(raw.ServiceName, raw.AggregateType)
	out := t.getDb().WithContext(ctx).
		Table(table).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}, {Name: "namespace"}},
			UpdateAll: true,
		}).
		Create(raw.Data)
	return out.Error
}

func (t *data) Load(ctx context.Context, serviceName string, aggregateName string, namespace string, id string, out interface{}) error {
	table := db.TableName(serviceName, aggregateName)

	r := t.getDb().WithContext(ctx).
		Table(table).
		Where("id = ?", id).
		Where("namespace = ?", namespace).
		First(out)
	return r.Error
}
func (t *data) Find(ctx context.Context, serviceName string, aggregateName string, namespace string, filter filters.Filter, out interface{}) error {
	table := db.TableName(serviceName, aggregateName)
	q := t.getDb().WithContext(ctx).
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
	var totalRows int64

	table := db.TableName(serviceName, aggregateName)
	q := t.getDb().WithContext(ctx).
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
