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

type data struct {
	db *gorm.DB
}

func (d *data) Initialize(cfg es.Config) error {
	if err := d.db.AutoMigrate(&event{}, &snapshot{}); err != nil {
		return err
	}

	entities := cfg.GetEntities()
	for _, raw := range entities {
		table := tableName(raw.ServiceName, raw.AggregateType)
		if err := d.db.Table(table).AutoMigrate(&entity{}); err != nil {
			return err
		}
		if err := d.db.Table(table).AutoMigrate(raw.Data); err != nil {
			return err
		}
	}
	return nil
}

func (d *data) NewTx(ctx context.Context) (es.Tx, error) {
	db := d.db.Begin()
	if db.Error != nil {
		return nil, db.Error
	}
	return newTransaction(db), nil
}

func NewData(dsn string) (es.Data, error) {
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

	d := &data{
		db: db,
	}
	return d, nil
}
