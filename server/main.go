package main

import (
	"log"
	"net"

	pb "github.com/contextcloud/eventstore/server/pb/store"

	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedStoreServer
}

func main() {
	listener, err := net.Listen("tcp", ":3332")
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	reflection.Register(s)

	server := &server{}

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(pb.Store_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)

	pb.RegisterStoreServer(s, server)
	healthpb.RegisterHealthServer(s, healthServer)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
