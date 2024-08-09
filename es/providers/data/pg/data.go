package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-apis/eventsourcing/es"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type lock func(ctx context.Context) error

func (l lock) Unlock(ctx context.Context) error {
	return l(ctx)
}

type data struct {
	service  string
	registry es.Registry
	db       *gorm.DB
	tx       *gorm.DB
}

func (d *data) getDb() *gorm.DB {
	if d.tx != nil {
		return d.tx
	}
	return d.db
}

func (d *data) Begin(ctx context.Context) (es.Tx, error) {
	_, span := otel.Tracer("local").Start(ctx, "Begin")
	defer span.End()

	if d.tx != nil {
		var rollback RollbackFunc
		var commitFunc CommitFunc

		if !d.tx.DisableNestedTransaction {
			// nested transaction
			//create save point
			spname := fmt.Sprintf("sp%p", d)
			if err := d.tx.SavePoint(spname).Error; err != nil {
				return nil, err
			}
			rollback = func() error {
				return d.tx.RollbackTo(spname).Error
			}
			//nested level do not need to commit
			commitFunc = func() error {
				return nil
			}
		}

		return &transaction{
			rollbackFunc: rollback,
			commitFunc:   commitFunc,
		}, nil
	}

	d.tx = d.db.Begin()

	var rollback = func() error {
		err := d.tx.Rollback().Error
		d.tx = nil
		return err
	}
	var commitFunc = func() error {
		err := d.tx.Commit().Error
		d.tx = nil
		return err
	}

	return &transaction{
		rollbackFunc: rollback,
		commitFunc:   commitFunc,
	}, d.tx.Error
}

func (d *data) Lock(ctx context.Context) (es.Lock, error) {
	_, span := otel.Tracer("local").Start(ctx, "Lock")
	defer span.End()

	db := d.getDb().WithContext(ctx).Exec("SELECT pg_advisory_lock(hashtext($1))", d.service)
	if db.Error != nil {
		return nil, db.Error
	}
	doit := func(ctx context.Context) error {
		return db.Exec("SELECT pg_advisory_unlock(hashtext($1))", d.service).Error
	}

	return lock(doit), nil
}

func (d *data) LoadSnapshot(ctx context.Context, search es.SnapshotSearch, out es.AggregateSourced) error {
	pctx, span := otel.Tracer("local").Start(ctx, "LoadSnapshot")
	defer span.End()

	var snapshot Snapshot
	r := d.getDb().
		WithContext(pctx).
		Model(&Snapshot{}).
		Where("service_name = ?", d.service).
		Where("namespace = ?", search.Namespace).
		Where("aggregate_type = ?", search.AggregateType).
		Where("aggregate_id = ?", search.AggregateId).
		Where("revision = ?", search.Revision).
		Limit(1).
		Find(&snapshot)
	if r.RowsAffected == 0 {
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

	if snapshot == nil {
		return nil // nothing to save
	}

	raw, err := json.Marshal(snapshot.Aggregate)
	if err != nil {
		return err
	}

	obj := &Snapshot{
		ServiceName:   d.service,
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

func (d *data) loadCommand(persisted *PersistedCommand) (es.Command, error) {
	commandConfig, err := d.registry.GetCommandConfig(persisted.Type)
	if err != nil {
		return nil, err
	}

	cmd, err := commandConfig.Factory()
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(persisted.Data, cmd); err != nil {
		return nil, err
	}

	return cmd, nil
}
func (d *data) SavePersistedCommand(ctx context.Context, cmd *es.PersistedCommand) error {
	pctx, span := otel.Tracer("local").Start(ctx, "SavePersistedCommand")
	defer span.End()

	raw, err := json.Marshal(cmd.Command)
	if err != nil {
		return err
	}

	obj := &PersistedCommand{
		ServiceName:  d.service,
		Namespace:    cmd.Namespace,
		Id:           cmd.Id,
		Type:         cmd.CommandType,
		Data:         raw,
		ExecuteAfter: cmd.ExecuteAfter,
		CreatedAt:    cmd.CreatedAt,
		By:           cmd.By,
	}

	out := d.getDb().
		WithContext(pctx).
		Clauses(clause.OnConflict{
			UpdateAll: true,
		}).
		Create(obj)
	return out.Error
}
func (d *data) DeletePersistedCommand(ctx context.Context, cmd *es.PersistedCommand) error {
	pctx, span := otel.Tracer("local").Start(ctx, "DeletePersistedCommand")
	defer span.End()

	out := d.getDb().
		WithContext(pctx).
		Delete(&PersistedCommand{
			ServiceName: d.service,
			Namespace:   cmd.Namespace,
			Id:          cmd.Id,
		})
	return out.Error
}
func (d *data) FindPersistedCommands(ctx context.Context, filter es.Filter) ([]*es.PersistedCommand, error) {
	pctx, span := otel.Tracer("local").Start(ctx, "FindPersistedCommands")
	defer span.End()

	q := d.getDb().
		WithContext(pctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Model(&PersistedCommand{}).
		Where("service_name = ?", d.service)

	if filter.Where != nil {
		q = where(q, filter.Where)
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
		q = q.Order(fmt.Sprintf("%s %s", order.Expression, strings.ToUpper(string(order.Direction))))
	}

	rows, err := q.
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cmds []*es.PersistedCommand
	for rows.Next() {
		var scanned PersistedCommand
		// ScanRows is a method of `gorm.DB`, it can be used to scan a row into a struct
		if err := q.ScanRows(rows, &scanned); err != nil {
			return nil, err
		}

		data, err := d.loadCommand(&scanned)
		if err != nil {
			return nil, err
		}

		cmds = append(cmds, &es.PersistedCommand{
			Id:           scanned.Id,
			Namespace:    scanned.Namespace,
			Command:      data,
			CommandType:  scanned.Type,
			ExecuteAfter: scanned.ExecuteAfter,
			CreatedAt:    scanned.CreatedAt,
			By:           scanned.By,
		})
	}

	return cmds, nil
}

func (d *data) NewScheduledCommandNotifier(ctx context.Context) (*es.ScheduledCommandNotifier, error) {
	inner, cancel := context.WithCancel(ctx)

	ch := make(chan time.Time)
	notifier := es.NewScheduledCommandNotifier(ch, func() {
		cancel()
		close(ch)
	})

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-inner.Done():
				return
			case t := <-ticker.C:
				ch <- t
			}
		}
	}()

	return notifier, nil
}

func (d *data) loadEventData(evt *Event) (interface{}, error) {
	eventConfig, err := d.registry.GetEventConfig(evt.ServiceName, evt.Type)
	if err != nil {
		return evt.Data, nil
	}

	data, err := eventConfig.Factory()
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(evt.Data, data); err != nil {
		return nil, err
	}

	return data, nil
}

func (d *data) FindEvents(ctx context.Context, filter es.Filter) ([]*es.Event, error) {
	pctx, span := otel.Tracer("local").Start(ctx, "GetEvents")
	defer span.End()

	q := d.getDb().
		WithContext(pctx).
		Model(&Event{}).
		Where("service_name = ?", d.service)

	if filter.Where != nil {
		q = where(q, filter.Where)
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
		q = q.Order(fmt.Sprintf("%s %s", order.Expression, strings.ToUpper(string(order.Direction))))
	}

	rows, err := q.
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*es.Event
	for rows.Next() {
		var evt Event
		// ScanRows is a method of `gorm.DB`, it can be used to scan a row into a struct
		if err := q.ScanRows(rows, &evt); err != nil {
			return nil, err
		}

		data, err := d.loadEventData(&evt)
		if err != nil {
			return nil, err
		}

		events = append(events, &es.Event{
			Service:       evt.ServiceName,
			Namespace:     evt.Namespace,
			AggregateId:   evt.AggregateId,
			AggregateType: evt.AggregateType,
			Type:          evt.Type,
			Version:       evt.Version,
			Timestamp:     evt.Timestamp,
			By:            evt.By,
			Data:          data,
			Metadata:      evt.Metadata,
		})
	}

	return events, nil
}
func (d *data) SaveEvents(ctx context.Context, events []*es.Event) error {
	pctx, span := otel.Tracer("local").Start(ctx, "SaveEvents")
	defer span.End()

	if len(events) == 0 {
		return nil // nothing to save
	}

	evts := make([]*Event, len(events))
	for i, evt := range events {
		raw, err := json.Marshal(evt.Data)
		if err != nil {
			return err
		}

		evts[i] = &Event{
			ServiceName:   d.service,
			Namespace:     evt.Namespace,
			AggregateId:   evt.AggregateId,
			AggregateType: evt.AggregateType,
			Type:          evt.Type,
			Version:       evt.Version,
			Timestamp:     evt.Timestamp,
			By:            evt.By,
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

	table := TableName(d.service, aggregateName)
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

	table := TableName(d.service, aggregateName)
	out := d.getDb().
		WithContext(pctx).
		Table(table).
		Delete(raw, "namespace = ?", raw.GetNamespace())
	return out.Error
}
func (d *data) Truncate(ctx context.Context, aggregateName string) error {
	pctx, span := otel.Tracer("local").Start(ctx, "Truncate")
	defer span.End()

	table := TableName(d.service, aggregateName)
	out := d.getDb().
		WithContext(pctx).
		Raw(fmt.Sprintf("TRUNCATE TABLE %s", table))
	return out.Error
}
func (d *data) Get(ctx context.Context, aggregateName string, namespace string, id uuid.UUID, out interface{}) error {
	pctx, span := otel.Tracer("local").Start(ctx, "Load")
	defer span.End()

	table := TableName(d.service, aggregateName)

	q := d.getDb().
		WithContext(pctx).
		Table(table).
		Where("id = ?", id)

	if namespace != "" {
		q = q.Where("namespace = ?", namespace)
	}

	r := q.Limit(1).Find(out)
	if r.RowsAffected == 0 {
		return sql.ErrNoRows
	}
	return r.Error
}
func (d *data) One(ctx context.Context, aggregateName string, namespace string, filter es.Filter, out interface{}) error {
	pctx, span := otel.Tracer("local").Start(ctx, "Load")
	defer span.End()

	table := TableName(d.service, aggregateName)

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

	r := q.Limit(1).Find(out)
	if r.RowsAffected == 0 {
		return sql.ErrNoRows
	}
	return r.Error
}
func (d *data) Find(ctx context.Context, aggregateName string, namespace string, filter es.Filter, out interface{}) error {
	pctx, span := otel.Tracer("local").Start(ctx, "Find")
	defer span.End()

	table := TableName(d.service, aggregateName)
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
		q = q.Order(fmt.Sprintf("%s %s", order.Expression, strings.ToUpper(string(order.Direction))))
	}

	r := q.
		Find(out)
	return r.Error
}
func (d *data) Count(ctx context.Context, aggregateName string, namespace string, filter es.Filter) (int, error) {
	pctx, span := otel.Tracer("local").Start(ctx, "Count")
	defer span.End()

	var totalRows int64

	table := TableName(d.service, aggregateName)
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

func newData(service string, db *gorm.DB, registry es.Registry) es.Data {
	d := &data{
		service:  service,
		db:       db,
		registry: registry,
	}
	return d
}
