package local

import (
	"context"
	"eventstore/es"
	"fmt"

	"github.com/uptrace/bun"
)

type postgresData struct {
	db *bun.DB
}

func (s *postgresData) GetTx(ctx context.Context) (es.Tx, error) {
	return es.TransactionCtx(ctx)
}

func (s *postgresData) WithTx(ctx context.Context) (context.Context, error) {
	// do we already have one?
	tx, _ := es.TransactionCtx(ctx)
	// return if we already have a transaction
	if tx != nil {
		return ctx, nil
	}

	// create a new one
	n, err := newPostgresTx(ctx, s)
	if err != nil {
		return nil, err
	}
	return es.SetTransaction(ctx, n), nil
}
func (s *postgresData) Load(ctx context.Context, id string, typeName string, out interface{}) error {
	tx, err := es.TransactionCtx(ctx)
	if err != nil {
		return err
	}
	if tx == nil {
		return es.ErrNoTransaction
	}
	return tx.Load(ctx, id, typeName, out)
}
func (s *postgresData) Save(ctx context.Context, id string, typeName string, out interface{}) ([]es.Event, error) {
	tx, err := es.TransactionCtx(ctx)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, es.ErrNoTransaction
	}
	return tx.Save(ctx, id, typeName, out)

}
func (s *postgresData) GetEvents(ctx context.Context) ([]es.Event, error) {
	// tx, err := es.TransactionCtx(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// if tx == nil {
	// 	return nil, es.ErrNoTransaction
	// }
	// return tx.GetEvents(ctx)
	return nil, fmt.Errorf("Not implemented")
}

func NewPostgresData(dsn string) (es.Data, error) {
	types := []interface{}{
		&dbEvent{},
		&dbSnapshot{},
	}
	factory, err := NewPostgresFactory(dsn, true, true, types)
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
