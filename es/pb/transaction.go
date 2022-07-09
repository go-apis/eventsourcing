package pb

import (
	"context"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/pb/store"
)

type transaction struct {
	storeClient store.StoreClient

	id string
}

func (t *transaction) Commit(ctx context.Context) (int, error) {
	resp, err := t.storeClient.Commit(ctx, &store.Tx{
		TransactionID: t.id,
	})
	if err != nil {
		return 0, err
	}

	return int(resp.UpdatedRows), nil
}

func (t *transaction) Rollback(ctx context.Context) error {
	_, err := t.storeClient.Rollback(ctx, &store.Tx{
		TransactionID: t.id,
	})
	if err != nil {
		return err
	}
	return nil
}

func newTransaction(storeClient store.StoreClient, id string) es.Tx {
	return &transaction{
		storeClient: storeClient,
		id:          id,
	}
}
