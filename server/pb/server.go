package pb

import (
	"context"
	"encoding/json"
	"time"

	"github.com/contextcloud/eventstore/pkg/db"
	"github.com/contextcloud/eventstore/server/pb/store"
	"github.com/contextcloud/eventstore/server/pb/streams"
	"github.com/contextcloud/eventstore/server/pb/transactions"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type dbEvent struct {
	ServiceName   string `gorm:"primaryKey"`
	Namespace     string `gorm:"primaryKey"`
	AggregateId   string `gorm:"primaryKey;type:uuid"`
	AggregateType string `gorm:"primaryKey"`
	Version       int    `gorm:"primaryKey"`
	Type          string `gorm:"primaryKey"`
	Timestamp     time.Time
	Data          json.RawMessage        `gorm:"serializer:json;type:jsonb"`
	Metadata      map[string]interface{} `gorm:"type:jsonb"`
}

type server struct {
	store.UnimplementedStoreServer

	gormDb              *gorm.DB
	transactionsManager transactions.Manager
	streamsManager      streams.Manager
}

func (s server) getDb(transactionId *string) (*gorm.DB, error) {
	if transactionId == nil {
		return s.gormDb, nil
	}

	transaction, err := s.transactionsManager.GetTransaction(*transactionId)
	if err != nil {
		return nil, err
	}

	return transaction.GetDb(), nil
}

func (s server) NewTx(ctx context.Context, req *store.NewTxRequest) (*store.NewTxResponse, error) {
	transaction, err := s.transactionsManager.NewTransaction(ctx)
	if err != nil {
		return nil, err
	}

	return &store.NewTxResponse{
		TransactionId: transaction.GetID(),
	}, nil
}

func (s server) Commit(ctx context.Context, tx *store.Tx) (*store.CommitResponse, error) {
	out, err := s.transactionsManager.CommitTransaction(tx.TransactionId)
	if err != nil {
		return nil, err
	}

	return &store.CommitResponse{
		UpdatedRows: int32(out),
	}, nil
}
func (s server) Rollback(ctx context.Context, tx *store.Tx) (*emptypb.Empty, error) {
	err := s.transactionsManager.DeleteTransaction(tx.TransactionId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s server) Events(ctx context.Context, req *store.EventsRequest) (*store.EventsResponse, error) {
	d, err := s.getDb(req.TransactionId)
	if err != nil {
		return nil, err
	}

	// do the query?
	var evts []*store.Event

	q := d.WithContext(ctx).
		Model(&db.Event{}).
		Order("version").
		Where("service_name = ?", req.ServiceName)

	if req.FromVersion != nil {
		q = q.Where("version >= ?", req.FromVersion)
	}
	if req.Namespace != nil {
		q = q.Where("namespace = ?", req.Namespace)
	}
	if req.AggregateId != nil {
		q = q.Where("aggregate_id = ?", req.AggregateId)
	}
	if req.AggregateType != nil {
		q = q.Where("aggregate_type = ?", req.AggregateType)
	}

	q = q.Scan(&evts)
	if q.Error != nil {
		return nil, q.Error
	}

	return &store.EventsResponse{
		Events: evts,
	}, nil
}

func (s server) SaveEvents(ctx context.Context, req *store.SaveEventsRequest) (*emptypb.Empty, error) {
	transaction, err := s.transactionsManager.GetTransaction(req.TransactionId)
	if err != nil {
		return nil, err
	}

	d := transaction.GetDb()
	q := d.WithContext(ctx).
		Create(&req.Events)

	if q.Error != nil {
		return nil, q.Error
	}
	return &emptypb.Empty{}, nil
}

func (s server) EventStream(stream store.Store_EventStreamServer) error {
	sender := s.streamsManager.NewStream(stream)
	defer s.streamsManager.DeleteSender(sender)

	if err := sender.Run(); err != nil {
		return err
	}
	return nil
}

func NewServer(gormDb *gorm.DB, transactionsManager transactions.Manager, streamsManager streams.Manager) store.StoreServer {
	return &server{
		gormDb:              gormDb,
		transactionsManager: transactionsManager,
		streamsManager:      streamsManager,
	}
}
