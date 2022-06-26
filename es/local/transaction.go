package local

import (
	"context"
	"encoding/json"

	"github.com/contextcloud/eventstore/es/filters"

	"github.com/contextcloud/eventstore/es"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type transaction struct {
	db *gorm.DB
}

func (t *transaction) LoadSnapshot(ctx context.Context, serviceName string, aggregateName string, namespace string, id string, out es.SourcedAggregate) error {
	return nil
}
func (t *transaction) GetEventDatas(ctx context.Context, serviceName string, aggregateName string, namespace string, id string, fromVersion int) ([]json.RawMessage, error) {
	var datas []json.RawMessage
	out := t.db.WithContext(ctx).
		Where("service_name = ?", serviceName).
		Where("namespace = ?", namespace).
		Where("aggregate_type = ?", aggregateName).
		Where("aggregate_id = ?", id).
		Where("version > ?", fromVersion).
		Order("version").
		Model(&event{}).
		Pluck("data", &datas)
	return datas, out.Error
}
func (t *transaction) SaveEvents(ctx context.Context, events []es.Event) error {
	if len(events) == 0 {
		return nil // nothing to save
	}

	data := make([]*event, len(events))
	for i, evt := range events {
		data[i] = &event{
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

	out := t.db.WithContext(ctx).
		Create(&data)
	return out.Error
}
func (t *transaction) SaveEntity(ctx context.Context, raw es.Entity) error {
	table := tableName(raw.ServiceName, raw.AggregateType)
	out := t.db.WithContext(ctx).
		Table(table).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}, {Name: "namespace"}},
			UpdateAll: true,
		}).
		Create(raw.Data)
	return out.Error
}
func (t *transaction) Commit(ctx context.Context) error {
	return t.db.Commit().Error
}
func (t *transaction) Rollback(ctx context.Context) error {
	return t.db.Rollback().Error
}

func (t *transaction) Load(ctx context.Context, serviceName string, aggregateName string, namespace string, id string, out interface{}) error {
	table := tableName(serviceName, aggregateName)

	r := t.db.WithContext(ctx).
		Table(table).
		Where("id = ?", id).
		Where("namespace = ?", namespace).
		First(out)
	return r.Error
}
func (t *transaction) Find(ctx context.Context, serviceName string, aggregateName string, namespace string, filter filters.Filter, out interface{}) error {
	table := tableName(serviceName, aggregateName)
	q := t.db.WithContext(ctx).
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
func (t *transaction) Count(ctx context.Context, serviceName string, aggregateName string, namespace string, filter filters.Filter) (int, error) {
	var totalRows int64

	table := tableName(serviceName, aggregateName)
	q := t.db.WithContext(ctx).
		Table(table).
		Where("namespace = ?", namespace)

	q = where(q, filter.Where)
	r := q.Count(&totalRows)
	return int(totalRows), r.Error
}

func newTransaction(db *gorm.DB) es.Tx {
	return &transaction{
		db: db,
	}
}
