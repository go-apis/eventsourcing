package local

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/filters"
	"github.com/contextcloud/eventstore/pkg/db"
	"github.com/google/uuid"

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

func (t *data) LoadSnapshot(ctx context.Context, serviceName string, aggregateName string, namespace string, revision string, id uuid.UUID, out es.AggregateSourced) error {
	return nil
}
func (d *data) SaveSnapshot(ctx context.Context, serviceName string, aggregateName string, namespace string, revision string, id uuid.UUID, out es.AggregateSourced) error {
	return nil
}

func (d *data) GetEventDatas(ctx context.Context, serviceName string, aggregateName string, namespace string, id uuid.UUID, fromVersion int) ([]*es.EventData, error) {
	var datas []*es.EventData
	out := d.getDb().
		WithContext(ctx).
		Where("service_name = ?", serviceName).
		Where("namespace = ?", namespace).
		Where("aggregate_type = ?", aggregateName).
		Where("aggregate_id = ?", id).
		Where("version > ?", fromVersion).
		Order("version").
		Table("events").
		Scan(&datas)

	return datas, out.Error
}
func (d *data) SaveEventDatas(ctx context.Context, serviceName string, aggregateName string, namespace string, id uuid.UUID, datas []*es.EventData) error {
	if !d.inTransaction() {
		return fmt.Errorf("must be in transaction")
	}

	if len(datas) == 0 {
		return nil // nothing to save
	}

	evts := make([]*db.Event, len(datas))
	for i, d := range datas {
		evts[i] = &db.Event{
			ServiceName:   serviceName,
			Namespace:     namespace,
			AggregateId:   id,
			AggregateType: aggregateName,
			Type:          d.Type,
			Version:       d.Version,
			Timestamp:     d.Timestamp,
			Data:          d.Data,
		}
	}

	out := d.getDb().
		WithContext(ctx).
		Create(&evts)
	return out.Error
}
func (t *data) SaveEntity(ctx context.Context, serviceName string, aggregateName string, raw es.Entity) error {
	if !t.inTransaction() {
		return fmt.Errorf("must be in transaction")
	}

	table := db.TableName(serviceName, aggregateName)
	out := t.getDb().
		WithContext(ctx).
		Table(table).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}, {Name: "namespace"}},
			UpdateAll: true,
		}).
		Create(raw)
	return out.Error
	return nil
}

func (t *data) Load(ctx context.Context, serviceName string, aggregateName string, namespace string, id uuid.UUID, out interface{}) error {
	table := db.TableName(serviceName, aggregateName)

	r := t.getDb().
		WithContext(ctx).
		Table(table).
		Where("id = ?", id).
		Where("namespace = ?", namespace).
		First(out)
	return r.Error
}
func (t *data) Find(ctx context.Context, serviceName string, aggregateName string, namespace string, filter filters.Filter, out interface{}) error {
	table := db.TableName(serviceName, aggregateName)
	q := t.getDb().
		WithContext(ctx).
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
