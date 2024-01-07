package pg

import (
	"context"

	"github.com/contextcloud/eventstore/es"
	"go.opentelemetry.io/otel"

	"gorm.io/gorm"
)

type conn struct {
	service  string
	registry es.Registry
	db       *gorm.DB
}

func (c *conn) initialize(ctx context.Context) error {
	_, pspan := otel.Tracer("local").Start(ctx, "Initialize")
	defer pspan.End()

	if err := c.db.AutoMigrate(&Event{}, &Snapshot{}); err != nil {
		return err
	}

	entities := c.registry.GetEntities()
	for _, opt := range entities {
		obj, err := opt.Factory()
		if err != nil {
			return err
		}

		table := TableName(c.service, opt.Name)
		if err := c.db.Table(table).AutoMigrate(&Entity{}); err != nil {
			return err
		}
		if err := c.db.Table(table).AutoMigrate(obj); err != nil {
			return err
		}
	}

	return nil
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
	c := &conn{
		service:  service,
		db:       db,
		registry: registry,
	}
	if err := c.initialize(ctx); err != nil {
		return nil, err
	}
	return c, nil
}
