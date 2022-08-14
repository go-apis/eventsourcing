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

func (c *conn) Initialize(ctx context.Context, initOpts es.InitializeOptions) (*es.Stream, error) {
	if err := c.db.AutoMigrate(&db.Event{}, &db.Snapshot{}); err != nil {
		return nil, err
	}

	for _, opt := range initOpts.EntityOptions {
		obj, err := opt.Factory()
		if err != nil {
			return nil, err
		}

		table := db.TableName(initOpts.ServiceName, opt.Name)
		if err := c.db.Table(table).AutoMigrate(&db.Entity{}); err != nil {
			return nil, err
		}
		if err := c.db.Table(table).AutoMigrate(obj); err != nil {
			return nil, err
		}
	}

	// todo create a stream
	stream := &es.Stream{}

	return stream, nil
}

func (c *conn) NewData(ctx context.Context) (es.Data, error) {
	db := c.db.WithContext(ctx)
	return newData(db), nil
}

func (c *conn) Publish(ctx context.Context, evts ...es.Event) error {
	return nil
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
