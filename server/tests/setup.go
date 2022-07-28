package tests

import (
	"github.com/contextcloud/eventstore/pkg/db"
	"github.com/contextcloud/eventstore/server/pb"
	store "github.com/contextcloud/eventstore/server/pb/store"
	"github.com/contextcloud/eventstore/server/pb/streams"
	"github.com/contextcloud/eventstore/server/pb/transactions"
	"gorm.io/gorm"
)

func CreateDb() (*gorm.DB, error) {
	return db.Open(
		db.WithDbUser("es"),
		db.WithDbPassword("es"),
		db.WithDbName("eventstore"),
		db.WithDebug(true),
	)
}

func CreateApiServer() (store.StoreServer, error) {
	gormDb, err := CreateDb()
	if err != nil {
		return nil, err
	}

	transactionOpts := transactions.DefaultOptions()
	transactionsManager, err := transactions.NewManager(transactionOpts, gormDb)
	if err != nil {
		return nil, err
	}

	streamsManager, err := streams.NewManager(gormDb)
	if err != nil {
		return nil, err
	}

	return pb.NewServer(gormDb, transactionsManager, streamsManager), nil
}
