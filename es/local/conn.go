package local

import (
	"context"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/pkg/db"

	"gorm.io/gorm"
)

type conn struct {
	db *gorm.DB
}

func (c *conn) Initialize(ctx context.Context, cfg es.Config) error {
	if err := c.db.AutoMigrate(&db.Event{}, &db.Snapshot{}); err != nil {
		return err
	}

	entities := cfg.GetEntities()
	for _, raw := range entities {
		table := db.TableName(raw.ServiceName, raw.AggregateType)
		if err := c.db.Table(table).AutoMigrate(&db.Entity{}); err != nil {
			return err
		}
		if err := c.db.Table(table).AutoMigrate(raw.Data); err != nil {
			return err
		}
	}
	return nil
}

func (c *conn) NewData(ctx context.Context) (es.Data, error) {
	db := c.db.WithContext(ctx)
	return newData(db), nil
}

func (c *conn) Close(ctx context.Context) error {
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
