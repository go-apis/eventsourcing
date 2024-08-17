package gdb

import (
	"context"

	"github.com/go-apis/eventsourcing/es"
	"go.opentelemetry.io/otel"

	"gorm.io/gorm"
)

func AutoMigrate(ctx context.Context, db *gorm.DB, service string, reg es.Registry) error {
	_, pspan := otel.Tracer("local").Start(ctx, "Initialize")
	defer pspan.End()

	if err := db.AutoMigrate(&Event{}, &Snapshot{}, &PersistedCommand{}); err != nil {
		return err
	}

	entities := reg.GetEntities()
	for _, opt := range entities {
		obj, err := opt.Factory()
		if err != nil {
			return err
		}

		table := TableName(service, opt.Name)
		if err := db.Table(table).AutoMigrate(&Entity{}); err != nil {
			return err
		}
		if err := db.Table(table).AutoMigrate(obj); err != nil {
			return err
		}
	}

	return nil
}

type conn struct {
	service  string
	registry es.Registry
	db       *gorm.DB
}

func (c *conn) NewData(ctx context.Context) (es.Data, error) {
	pctx, pspan := otel.Tracer("local").Start(ctx, "NewData")
	defer pspan.End()

	db := c.db.WithContext(pctx)
	return newData(c.service, db, c.registry), nil
}

func (c *conn) Close(ctx context.Context) error {
	_, pspan := otel.Tracer("local").Start(ctx, "Close")
	defer pspan.End()

	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func NewConn(ctx context.Context, service string, db *gorm.DB, registry es.Registry) (es.Conn, error) {
	return &conn{
		service:  service,
		db:       db,
		registry: registry,
	}, nil
}
