package tests

import (
	context "context"
	"errors"

	store "github.com/contextcloud/eventstore/server/pb/store"
	_ "github.com/golang/mock/mockgen/model"
	"google.golang.org/grpc"
)

func NewStreamMock() *StreamMock {
	return &StreamMock{
		ctx:            context.Background(),
		recvToServer:   make(chan *store.EventStreamRequest, 10),
		sentFromServer: make(chan *store.EventStreamResponse, 10),
	}
}

type StreamMock struct {
	grpc.ServerStream
	ctx            context.Context
	recvToServer   chan *store.EventStreamRequest
	sentFromServer chan *store.EventStreamResponse
}

func (m *StreamMock) Context() context.Context {
	return m.ctx
}
func (m *StreamMock) Send(resp *store.EventStreamResponse) error {
	m.sentFromServer <- resp
	return nil
}
func (m *StreamMock) Recv() (*store.EventStreamRequest, error) {
	req, more := <-m.recvToServer
	if !more {
		return nil, errors.New("empty")
	}
	return req, nil
}
func (m *StreamMock) SendFromClient(req *store.EventStreamRequest) error {
	m.recvToServer <- req
	return nil
}
func (m *StreamMock) RecvToClient() (*store.EventStreamResponse, error) {
	response, more := <-m.sentFromServer
	if !more {
		return nil, errors.New("empty")
	}
	return response, nil
}
