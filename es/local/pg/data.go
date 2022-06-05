package pg

import (
	"context"
	"encoding/json"
	"eventstore/es"

	"github.com/uptrace/bun"
)

type postgresData struct {
	db *bun.DB
}

func (s *postgresData) NewTx(ctx context.Context) (es.Tx, error) {
	return newTx(ctx, s)
}

func (s *postgresData) LoadSnapshot(ctx context.Context, serviceName string, aggregateName string, namespace string, id string, out es.SourcedAggregate) error {
	tx, err := transactionCtx(ctx)
	if err != nil {
		return err
	}
	if tx == nil {
		return ErrNoTransaction
	}
	return tx.LoadSnapshot(ctx, serviceName, aggregateName, namespace, id, out)
}
func (s *postgresData) GetEventDatas(ctx context.Context, serviceName string, aggregateName string, namespace string, id string, fromVersion int) ([]json.RawMessage, error) {
	tx, err := transactionCtx(ctx)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, ErrNoTransaction
	}
	return tx.GetEventDatas(ctx, serviceName, aggregateName, namespace, id, fromVersion)
}
func (s *postgresData) SaveEvents(ctx context.Context, events []es.Event) error {
	tx, err := transactionCtx(ctx)
	if err != nil {
		return err
	}
	if tx == nil {
		return ErrNoTransaction
	}
	return tx.SaveEvents(ctx, events)
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
