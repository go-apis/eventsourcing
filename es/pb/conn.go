package pb

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/pb/store"
	multierror "github.com/hashicorp/go-multierror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type conn struct {
	clientConn  *grpc.ClientConn
	storeClient store.StoreClient
	streamer    Streamer
}

func (c *conn) Initialize(ctx context.Context, cfg es.Config) error {
	// todo send schemas to server

	// subscribe to events
	// c.streamer = NewStreamer(c.storeClient, cfg)
	// if err := c.streamer.Run(ctx); err != nil {
	// 	return err
	// }
	return nil
}

func (c *conn) NewData(ctx context.Context) (es.Data, error) {
	return newData(c.storeClient)
}

func (c *conn) Close(ctx context.Context) error {
	var result error

	if err := c.clientConn.Close(); err != nil {
		result = multierror.Append(result, err)
	}
	if c.streamer != nil {
		if err := c.streamer.Close(ctx); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

func NewConn(dsn string) (es.Conn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	// todo..
	ctx := context.Background()

	c, err := grpc.DialContext(ctx, dsn, opts...)
	if err != nil {
		return nil, err
	}

	healthClient := healthpb.NewHealthClient(c)
	storeClient := store.NewStoreClient(c)

	resp, err := healthClient.Check(ctx, &healthpb.HealthCheckRequest{
		Service: store.Store_ServiceDesc.ServiceName,
	})
	if err != nil {
		return nil, err
	}
	if resp.Status != healthpb.HealthCheckResponse_SERVING {
		return nil, fmt.Errorf("service %s status: %s", store.Store_ServiceDesc.ServiceName, resp.Status)
	}

	return &conn{
		clientConn:  c,
		storeClient: storeClient,
	}, nil
}
