package main

import (
	"context"
	pb "eventstore/server/pb/store"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedStoreServer
}

func (s *server) Watch(in *pb.WatchRequest, srv pb.Store_WatchServer) error {
	log.Printf("fetch response for id : %v", in)
	//use wait group to allow process to be concurrent
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(count int64) {
			defer wg.Done()

			//time sleep to simulate server process time
			time.Sleep(time.Duration(count) * time.Second)
			resp := pb.WatchResponse{}
			if err := srv.Send(&resp); err != nil {
				log.Printf("send error %v", err)
			}
			log.Printf("finishing request number : %d", count)
		}(int64(i))
	}

	wg.Wait()
	return nil
}

func (s *server) GetEvent(ctx context.Context, in *pb.GetEventRequest) (*pb.GetEventResponse, error) {
	log.Printf("Received request: %v", in.ProtoReflect().Descriptor().FullName())

	evt := &pb.Event{
		Title:     "The Hitchhiker's Guide to the Galaxy",
		Author:    "Douglas Adams",
		PageCount: 42,
	}
	return &pb.GetEventResponse{
		Event: evt,
	}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	pb.RegisterStoreServer(s, &server{})
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
