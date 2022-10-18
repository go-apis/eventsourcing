package core

import (
	"context"

	"github.com/contextcloud/eventstore/server/pb/streams"
	"github.com/contextcloud/graceful/srv"
)

type streamsServer struct {
	sm streams.Manager
}

func (s *streamsServer) Start(ctx context.Context) error {
	return nil
}

func (s *streamsServer) Shutdown(ctx context.Context) error {
	return s.sm.Stop()
}

func NewStreamsServer(sm streams.Manager) (srv.Startable, error) {
	return &streamsServer{
		sm: sm,
	}, nil
}
