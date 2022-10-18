package core

import (
	"context"

	"github.com/contextcloud/eventstore/server/pb/transactions"
	"github.com/contextcloud/graceful/srv"
)

type transactionsServer struct {
	tm transactions.Manager
}

func (s *transactionsServer) Start(ctx context.Context) error {
	return s.tm.Start()
}

func (s *transactionsServer) Shutdown(ctx context.Context) error {
	return s.tm.Stop()
}

func NewTransactionsServer(tm transactions.Manager) (srv.Startable, error) {
	return &transactionsServer{
		tm: tm,
	}, nil
}
