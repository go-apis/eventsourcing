package pb

import (
	"context"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/pb/store"
	"google.golang.org/grpc"
)

type data struct {
	storeClient store.StoreClient
}

func (d *data) NewTx(ctx context.Context) (es.Tx, error) {
	tx, err := d.storeClient.NewTx(ctx)

	return nil, nil
}

func NewData(dns string) (es.Data, error) {
	var opts []grpc.DialOption
	conn, err := grpc.Dial("localhost:6565", opts...)
	if err != nil {
		return nil, err
	}

	storeClient := store.NewStoreClient(conn)

	return &data{
		storeClient: storeClient,
	}, nil
}
