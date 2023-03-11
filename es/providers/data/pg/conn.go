package pg

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/pkg/pgdb"
	"go.opentelemetry.io/otel"

	"gorm.io/gorm"
)

type conn struct {
	initialized bool
	serviceName string
	db          *gorm.DB
}

func (c *conn) Initialize(ctx context.Context, initOpts es.InitializeOptions) error {
	_, pspan := otel.Tracer("local").Start(ctx, "Initialize")
	defer pspan.End()

	if err := c.db.AutoMigrate(&pgdb.Event{}, &pgdb.Snapshot{}); err != nil {
		return err
	}

	for _, opt := range initOpts.EntityConfigs {
		obj, err := opt.Factory()
		if err != nil {
			return err
		}

		table := pgdb.TableName(initOpts.ServiceName, opt.Name)
		if err := c.db.Table(table).AutoMigrate(&pgdb.Entity{}); err != nil {
			return err
		}
		if err := c.db.Table(table).AutoMigrate(obj); err != nil {
			return err
		}
	}

	c.initialized = true
	c.serviceName = initOpts.ServiceName
	return nil
}

func (c *conn) NewData(ctx context.Context) (es.Data, error) {
	pctx, pspan := otel.Tracer("local").Start(ctx, "NewData")
	defer pspan.End()

	if !c.initialized {
		return nil, fmt.Errorf("conn not initialized")
	}

	db := c.db.WithContext(pctx)
	return newData(c.serviceName, db), nil
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

func NewConn(db *gorm.DB) (es.Conn, error) {
	return &conn{db: db}, nil
}
