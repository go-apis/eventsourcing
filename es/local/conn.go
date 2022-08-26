package local

import (
	"context"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/pkg/db"
	"go.opentelemetry.io/otel"

	"gorm.io/gorm"
)

type conn struct {
	db *gorm.DB
}

func (c *conn) Initialize(ctx context.Context, initOpts es.InitializeOptions) error {
	_, pspan := otel.Tracer("local").Start(ctx, "Initialize")
	defer pspan.End()

	if err := c.db.AutoMigrate(&db.Event{}, &db.Snapshot{}); err != nil {
		return err
	}

	for _, opt := range initOpts.EntityOptions {
		obj, err := opt.Factory()
		if err != nil {
			return err
		}

		table := db.TableName(initOpts.ServiceName, opt.Name)
		if err := c.db.Table(table).AutoMigrate(&db.Entity{}); err != nil {
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
	return newData(db), nil
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

func NewConn(opts ...db.OptionFunc) (es.Conn, error) {
	gormDb, err := db.Open(opts...)
	if err != nil {
		return nil, err
	}

	c := &conn{
		db: gormDb,
	}
	return c, nil
}
