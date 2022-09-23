package server

import (
	"context"

	"github.com/contextcloud/eventstore/pkg/db"
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

type ServerDbConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	Debug    bool
}

type ServerConfig struct {
	Db ServerDbConfig
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Db: ServerDbConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "es",
			Password: "es",
			Name:     "eventstore",
			Debug:    true,
		},
	}
}

func NewServer(ctx context.Context, cfg *config.Config) (srv.Startable, error) {
	srvCfg := NewServerConfig()
	if err := cfg.Parse(srvCfg); err != nil {
		return nil, err
	}

	gormDb, err := db.Open(
		db.WithDbHost(srvCfg.Db.Host),
		db.WithDbPort(srvCfg.Db.Port),
		db.WithDbUser(srvCfg.Db.User),
		db.WithDbPassword(srvCfg.Db.Password),
		db.WithDbName(srvCfg.Db.Name),
		db.WithDebug(srvCfg.Db.Debug),
	)
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
