package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/contextcloud/eventstore/pkg/db"
	"github.com/contextcloud/eventstore/server/pb"
	"github.com/contextcloud/eventstore/server/pb/logger"
	"github.com/contextcloud/eventstore/server/pb/store"
	"github.com/contextcloud/eventstore/server/pb/streams"
	"github.com/contextcloud/eventstore/server/pb/transactions"

	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logit := logger.NewSimpleLogger("server ", os.Stderr)

	listener, err := net.Listen("tcp", ":3332")
	if err != nil {
		panic(err)
	}
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(store.Store_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)

	gormDb, err := db.Open(
		db.WithDbUser("es"),
		db.WithDbPassword("es"),
		db.WithDbName("eventstore"),
		db.WithDebug(true),
	)
	if err != nil {
		logit.Errorf("failed to open db: %v", err)
		return
	}

	transactionOpts := transactions.DefaultOptions()
	transactionsManager, err := transactions.NewManager(transactionOpts, gormDb)
	if err != nil {
		logit.Errorf("failed to create transactions manager: %v", err)
		return
	}
	streamsManager, err := streams.NewManager(gormDb)
	if err != nil {
		logit.Errorf("failed to create streams manager: %v", err)
		return
	}

	srv := pb.NewServer(gormDb, transactionsManager, streamsManager)

	s := grpc.NewServer()
	reflection.Register(s)
	store.RegisterStoreServer(s, srv)
	healthpb.RegisterHealthServer(s, healthServer)

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := transactionsManager.Start(); err != nil {
		logit.Errorf("failed to start transactions manager: %v", err)
		log.Fatal(err)
		return
	}

	go func() {
		if err := s.Serve(listener); err != nil {
			logit.Errorf("failed to serve: %v", err)
			log.Fatal(err)
			return
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	logit.Infof("Caught SIGTERM")

	if err := transactionsManager.Stop(); err != nil {
		logit.Errorf("failed to stop transactions manager: %v", err)
	}
	if err := streamsManager.Stop(); err != nil {
		logit.Errorf("failed to stop streams manager: %v", err)
	}

	s.GracefulStop()

	logit.Infof("Shutdown completed")
}
