package streams

import (
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/contextcloud/eventstore/server/pb/store"
)

type Sender interface {
	Run() error
}

type sender struct {
	mux sync.RWMutex

	stream store.Store_EventStreamServer
}

func (s *sender) handleInit(req *store.EventStreamRequest_Init) error {
	return nil
}

func (s *sender) handleNack(req *store.EventId) error {
	return nil
}

func (s *sender) handleAck(req *store.EventId) error {
	return nil
}

func (s *sender) handle(req *store.EventStreamRequest) error {
	switch msg := req.Msg.(type) {
	case *store.EventStreamRequest_InitMsg:
		return s.handleInit(msg.InitMsg)
	case *store.EventStreamRequest_NackMsg:
		return s.handleNack(msg.NackMsg)
	case *store.EventStreamRequest_AckMsg:
		return s.handleAck(msg.AckMsg)
	default:
		return fmt.Errorf("unknown message type %T", msg)
	}
}

func (s *sender) Run() error {
	// this should be a for loop
	ctx := s.stream.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			evt, err := s.stream.Recv()
			if err == io.EOF {
				// return will close stream from server side
				log.Println("exit")
				// todo remove from parent.
				return nil
			}
			if err != nil {
				log.Printf("receive error %v", err)
				continue
			}

			// handle the message
			if err := s.handle(evt); err != nil {
				log.Printf("handle error %v", err)
				continue
			}
		}
	}

	return nil
}

func NewSender(stream store.Store_EventStreamServer) Sender {
	return &sender{
		stream: stream,
	}
}
