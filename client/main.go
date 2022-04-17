package main

import (
	"context"
	pb "eventstore/client/pb/store"
	"io"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewStoreClient(conn)
	bookList, err := client.GetEvent(context.Background(), &pb.GetEventRequest{})
	if err != nil {
		log.Fatalf("failed to get book list: %v", err)
		return
	}
	log.Printf("book list: %v", bookList)

	ctx := context.Background()

	in := &pb.WatchRequest{}
	stream, err := client.Watch(ctx, in)
	if err != nil {
		log.Fatalf("open stream error %v", err)
		return
	}

	done := make(chan bool)
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				done <- true //means stream is finished
				return
			}
			if err != nil {
				log.Fatalf("cannot receive %v", err)
			}
			log.Printf("Resp received: %s", resp.String())
		}
	}()

	<-done //we will wait until all response is received
	log.Printf("finished")
}
