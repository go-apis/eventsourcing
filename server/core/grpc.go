package core

import (
	"context"
	"net"

	"github.com/contextcloud/graceful/srv"
	"google.golang.org/grpc"
)

type grpcServer struct {
	listener net.Listener
}

func (g *grpcServer) Start(ctx context.Context) error {

	return nil
}

func (g *grpcServer) Shutdown(ctx context.Context) error {
	return nil
}

func NewGrpcServer(gs *grpc.Server, address string) (srv.Startable, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	return &grpcServer{
		listener: listener,
	}, nil
}
