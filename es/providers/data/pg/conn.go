package pg

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"go.opentelemetry.io/otel"

	"gorm.io/gorm"
)

type conn struct {
	initialized bool
	service     string
	db          *gorm.DB
}

func (c *conn) Initialize(ctx context.Context, cfg es.Config) error {
	_, pspan := otel.Tracer("local").Start(ctx, "Initialize")
	defer pspan.End()

	if err := c.db.AutoMigrate(&Event{}, &Snapshot{}); err != nil {
		return err
	}

	service := cfg.
		GetProviderConfig().
		Service

	for _, opt := range cfg.GetEntityConfigs() {
		obj, err := opt.Factory()
		if err != nil {
			return err
		}

		table := TableName(service, opt.Name)
		if err := c.db.Table(table).AutoMigrate(&Entity{}); err != nil {
			return err
		}
		if err := c.db.Table(table).AutoMigrate(obj); err != nil {
			return err
		}
	}

	c.initialized = true
	c.service = service
	return nil
}

func (c *conn) NewData(ctx context.Context) (es.Data, error) {
	pctx, pspan := otel.Tracer("local").Start(ctx, "NewData")
	defer pspan.End()

	if !c.initialized {
		return nil, fmt.Errorf("conn not initialized")
	}

	db := c.db.WithContext(pctx)
	return newData(c.service, db), nil
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
