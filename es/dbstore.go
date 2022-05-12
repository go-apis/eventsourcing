package es

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

type dbEvent struct {
	bun.BaseModel `bun:"table:events,alias:evt"`

	ServiceName   string `bun:",pk"`
	Namespace     string `bun:",pk"`
	AggregateId   string `bun:",pk,type:uuid"`
	AggregateType string `bun:",pk"`
	Version       int    `bun:",pk"`
	Type          string `bun:",notnull"`
	Timestamp     time.Time
	Data          json.RawMessage `bun:"type:jsonb"`
	Metadata      Metadata        `bun:"type:jsonb"`
}

// todo add version
type dbSnapshot struct {
	bun.BaseModel `bun:"table:events,alias:evt"`

	ServiceName string `bun:",pk"`
	Namespace   string `bun:",pk"`
	Id          string `bun:",pk,type:uuid"`
	Type        string `bun:",pk"`
	Revision    string `bun:",pk"`
	Aggregate   []byte `bun:",notnull,type:jsonb"`
}

type dbStore struct {
	serviceName string
	db          *bun.DB
	tx          *bun.Tx
}

func (s *dbStore) idb() bun.IDB {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

func (s *dbStore) loadSnapshot(ctx context.Context, namespace string, id string, typeName string, out interface{}) (int, error) {
	return 0, nil
}
func (s *dbStore) loadEvents(ctx context.Context, namespace string, id string, typeName string, from int) ([]dbEvent, error) {
	// Select all users.
	var evts []dbEvent
	if err := s.idb().NewSelect().
		Model(&evts).
		Where("service_name = ?", s.serviceName).
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

func (s *dbStore) loadSourced(ctx context.Context, id string, typeName string, out SourcedAggregate) error {
	namespace := NamespaceFromContext(ctx)

	// load from snapshot
	version, err := s.loadSnapshot(ctx, namespace, id, typeName, out)
	if err != nil {
		return err
	}

	// get the events
	events, err := s.loadEvents(ctx, namespace, id, typeName, version)
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
func (s *dbStore) saveSourced(ctx context.Context, id string, typeName string, out SourcedAggregate) ([]Event, error) {
	datas := out.GetEvents()
	if len(datas) == 0 {
		return nil, nil // nothing to save
	}

	namespace := NamespaceFromContext(ctx)
	version := out.GetVersion()

	// get the events
	evts := make([]Event, len(datas))
	dbEvts := make([]dbEvent, len(datas))
	for i, data := range datas {
		buf, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		name := fmt.Sprintf("%T", data)
		metadata := MetadataFromContext(ctx)
		v := version + i + 1
		ts := time.Now()

		evts[i] = Event{
			ServiceName:   s.serviceName,
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
			ServiceName:   s.serviceName,
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
	if _, err := s.idb().
		NewInsert().
		Model(&dbEvts).
		Exec(ctx); err != nil {
		return nil, err
	}
	return evts, nil
}

func (s *dbStore) NewTx(ctx context.Context) (Tx, error) {
	c := s.tx
	if c == nil {
		// create one then return
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		c = &tx
	}

	n := &dbStore{
		serviceName: s.serviceName,
		db:          s.db,
		tx:          c,
	}
	return NewTx(ctx, n, c.Commit)
}
func (s *dbStore) Load(ctx context.Context, id string, typeName string, out interface{}) error {
	switch impl := out.(type) {
	case SourcedAggregate:
		return s.loadSourced(ctx, id, typeName, impl)
	default:
		return fmt.Errorf("Invalid aggregate type")
	}
}
func (s *dbStore) Save(ctx context.Context, id string, typeName string, out interface{}) ([]Event, error) {
	switch impl := out.(type) {
	case SourcedAggregate:
		return s.saveSourced(ctx, id, typeName, impl)
	default:
		return nil, fmt.Errorf("Invalid aggregate type")
	}
}
func (s *dbStore) GetEvents(ctx context.Context) ([]Event, error) {
	// Select all users.
	var evts []Event
	if err := s.idb().NewSelect().
		Model(&evts).
		Order("timestamp desc").
		Scan(ctx); err != nil {
		if err != nil && sql.ErrNoRows != err {
			return nil, err
		}
	}
	return evts, nil
}

func NewDbStore(dsn string, serviceName string) (Store, error) {
	types := []interface{}{
		&dbEvent{},
		&dbSnapshot{},
	}
	factory, err := NewDbFactory(dsn, true, true, types)
	if err != nil {
		return nil, err
	}
	db, err := factory.New()
	if err != nil {
		return nil, err
	}

	return &dbStore{
		serviceName: serviceName,
		db:          db,
	}, nil
}
