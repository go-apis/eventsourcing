package pb

import (
	"context"

	"github.com/contextcloud/eventstore/server/pb/store"
	"github.com/contextcloud/eventstore/server/pb/transactions"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	store.UnimplementedStoreServer

	transactionsManager transactions.Manager
}

func (s server) NewTx(context.Context, *store.NewTxRequest) (*store.NewTxResponse, error) {
	s.transactionsManager.NewSession()

	return nil, status.Errorf(codes.Unimplemented, "method NewTx not implemented")
}

func NewServer(transactionsManager transactions.Manager) store.StoreServer {
	return &server{
		transactionsManager: transactionsManager,
	}
}
