package main

import (
	"log"
	"net"

	"github.com/contextcloud/eventstore/server/pb"
	"github.com/contextcloud/eventstore/server/pb/store"

	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	listener, err := net.Listen("tcp", ":3332")
	if err != nil {
		panic(err)
	}

	srv := pb.NewServer()

	s := grpc.NewServer()
	reflection.Register(s)

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(store.Store_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)

	store.RegisterStoreServer(s, srv)
	healthpb.RegisterHealthServer(s, healthServer)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
