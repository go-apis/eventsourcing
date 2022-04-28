package es

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ugorji/go/codec"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type dbEvent struct {
	Namespace     string `bun:",pk"`
	AggregateId   string `bun:",pk,type:uuid"`
	AggregateType string `bun:",pk"`
	Version       int    `bun:",pk"`
	Type          string `bun:",notnull"`
	Timestamp     time.Time
	Data          json.RawMessage        `bun:"type:jsonb"`
	Metadata      map[string]interface{} `bun:"type:jsonb"`
}

// todo add version
type dbSnapshot struct {
	Namespace string          `bun:",pk"`
	Id        string          `bun:",pk,type:uuid"`
	Type      string          `bun:",pk"`
	Revision  string          `bun:",pk"`
	Aggregate json.RawMessage `bun:",notnull,type:jsonb"`
}

type dbStore struct {
	db *bun.DB
	ch codec.Handle
}

func (db *dbStore) loadSnapshot(ctx context.Context, namespace string, id string, typeName string, out interface{}) (int, error) {
	return 0, nil
}

func (db *dbStore) loadEvents(ctx context.Context, namespace string, id string, typeName string, from int) ([]dbEvent, error) {
	return nil, nil
}

func (db *dbStore) loadSourced(ctx context.Context, id string, typeName string, out SourcedAggregate) error {
	namespace := "default"

	// load from snapshot
	version, err := db.loadSnapshot(ctx, namespace, id, typeName, out)
	if err != nil {
		return err
	}

	// get the events
	events, err := db.loadEvents(ctx, namespace, id, typeName, version)
	if err != nil {
		return err
	}

	for _, evt := range events {
		r := bytes.NewReader(evt.Data)
		d := codec.NewDecoder(r, db.ch)

		if err := d.Decode(out); err != nil {
			return err
		}
		out.IncrementVersion()
	}

	return nil
}

func (db *dbStore) Load(ctx context.Context, id string, typeName string, out interface{}) error {
	switch impl := out.(type) {
	case SourcedAggregate:
		return db.loadSourced(ctx, id, typeName, impl)
	default:
		return fmt.Errorf("Invalid aggregate type")
	}

	return nil
}

func NewDbStore(dsn string) (Store, error) {
	var jh codec.JsonHandle

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	return &dbStore{
		db: db,
		ch: &jh,
	}, nil
}
