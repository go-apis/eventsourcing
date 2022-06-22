package g

import (
	"context"
	"log"
	"os"
	"time"

	"eventstore/es"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type data struct {
	db *gorm.DB
}

func (d *data) NewTx(ctx context.Context) (es.Tx, error) {
	db := d.db.Begin()
	if db.Error != nil {
		return nil, db.Error
	}
	return newTransaction(db), nil
}

func NewData(cfg es.Config, dsn string) (es.Data, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Disable color
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&event{}, &snapshot{}); err != nil {
		return nil, err
	}

	entities := cfg.GetEntities()
	for _, raw := range entities {
		table := tableName(raw.ServiceName, raw.AggregateType)
		if err := db.Table(table).AutoMigrate(&entity{}); err != nil {
			return nil, err
		}
		if err := db.Table(table).AutoMigrate(raw.Data); err != nil {
			return nil, err
		}
	}

	d := &data{
		db: db,
	}
	return d, nil
}
