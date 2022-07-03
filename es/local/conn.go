package local

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/contextcloud/eventstore/es"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type conn struct {
	db *gorm.DB
}

func (c *conn) Initialize(ctx context.Context, cfg es.Config) error {
	if err := c.db.AutoMigrate(&event{}, &snapshot{}); err != nil {
		return err
	}

	entities := cfg.GetEntities()
	for _, raw := range entities {
		table := tableName(raw.ServiceName, raw.AggregateType)
		if err := c.db.Table(table).AutoMigrate(&entity{}); err != nil {
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

func NewConn(dsn string) (es.Conn, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Disable color
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 newLogger,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}

	c := &conn{
		db: db,
	}
	return c, nil
}
