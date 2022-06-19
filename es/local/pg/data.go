package pg

import (
	"context"
	"eventstore/es"

	"github.com/uptrace/bun"
)

type postgresData struct {
	db *bun.DB
}

func (s *postgresData) NewTx(ctx context.Context) (es.Tx, error) {
	return newTx(ctx, s)
}

func NewData(dsn string) (es.Data, error) {
	types := []interface{}{
		&dbEvent{},
		&dbSnapshot{},
	}
	factory, err := newFactory(dsn, true, true, types)
	if err != nil {
		return nil, err
	}
	db, err := factory.New()
	if err != nil {
		return nil, err
	}

	return &postgresData{
		db: db,
	}, nil
}
