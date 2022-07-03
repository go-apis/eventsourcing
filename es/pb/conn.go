package pb

import (
	"context"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/pb/store"
	"google.golang.org/grpc"
)

type conn struct {
	storeClient store.StoreClient
}

func (c *conn) NewData(ctx context.Context) (es.Data, error) {
	return newData(c.storeClient)
}

func (c *conn) Initialize(ctx context.Context, cfg es.Config) error {
	return nil
}

func NewConn(dsn string) (es.Conn, error) {
	var opts []grpc.DialOption
	c, err := grpc.Dial(dsn, opts...)
	if err != nil {
		return nil, err
	}

	storeClient := store.NewStoreClient(c)

	return &conn{
		storeClient: storeClient,
	}, nil
}
