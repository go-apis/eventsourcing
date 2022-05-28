package dg

import (
	"context"
	"eventstore/es"

	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type data struct {
	cli *dgo.Dgraph
}

func (d *data) WithTx(ctx context.Context) (context.Context, es.Tx, error) {
	// tx := d.cli.NewTxn()
	// tx.Commit()

	return ctx, nil, nil
}
func (d *data) Load(ctx context.Context, serviceName string, aggregateName string, id string, out interface{}) error {
	return nil
}
func (d *data) Save(ctx context.Context, serviceName string, aggregateName string, id string, val interface{}) ([]es.Event, error) {
	return nil, nil
}

func NewData(dsn string) (es.Data, error) {
	// Dial a gRPC connection. The address to dial to can be configured when
	// setting up the dgraph cluster.

	dialOpts := append([]grpc.DialOption{},
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	d, err := grpc.Dial("localhost:9080", dialOpts...)
	if err != nil {
		return nil, err
	}
	cli := dgo.NewDgraphClient(
		api.NewDgraphClient(d),
	)

	return &data{
		cli: cli,
	}, nil
}
