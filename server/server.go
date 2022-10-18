package server

import (
	"context"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/pkg/pgdb"
	"github.com/contextcloud/eventstore/server/core"
	"github.com/contextcloud/eventstore/server/pb"
	"github.com/contextcloud/eventstore/server/pb/store"
	"github.com/contextcloud/eventstore/server/pb/streams"
	"github.com/contextcloud/eventstore/server/pb/transactions"
	"github.com/contextcloud/graceful/config"
	"github.com/contextcloud/graceful/srv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func NewServer(ctx context.Context, cfg *config.Config) (srv.Startable, error) {
	pcfg := &es.ProviderConfig{}
	if err := cfg.Parse(pcfg); err != nil {
		return nil, err
	}

	gormDb, err := pgdb.Open(pcfg.Data.Pg)
	if err != nil {
		return nil, err
	}

	transactionOpts := transactions.DefaultOptions()
	transactionsManager, err := transactions.NewManager(transactionOpts, gormDb)
	if err != nil {
		return nil, err
	}
	streamsManager, err := streams.NewManager(gormDb)
	if err != nil {
		return nil, err
	}
	s := pb.NewServer(gormDb, transactionsManager, streamsManager)

	healthServer := health.NewServer()
	healthServer.SetServingStatus(cfg.ServiceName, healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(store.Store_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)

	gs := grpc.NewServer()
	reflection.Register(gs)
	store.RegisterStoreServer(gs, s)
	healthpb.RegisterHealthServer(gs, healthServer)

	grpcServer, err := core.NewGrpcServer(gs, cfg.SrvAddr)
	if err != nil {
		return nil, err
	}
	streamsServer, err := core.NewStreamsServer(streamsManager)
	if err != nil {
		return nil, err
	}
	transactionsServer, err := core.NewTransactionsServer(transactionsManager)
	if err != nil {
		return nil, err
	}

	multi := srv.NewMulti(
		grpcServer,
		streamsServer,
		transactionsServer,
	)
	return multi, nil
}
