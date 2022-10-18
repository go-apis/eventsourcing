package tests

import (
	"github.com/contextcloud/eventstore/pkg/pgdb"
	"github.com/contextcloud/eventstore/server/pb"
	store "github.com/contextcloud/eventstore/server/pb/store"
	"github.com/contextcloud/eventstore/server/pb/streams"
	"github.com/contextcloud/eventstore/server/pb/transactions"
	"gorm.io/gorm"
)

func CreateDb() (*gorm.DB, error) {
	cfg := &pgdb.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "es",
		Password: "es",
		Name:     "eventstore",
		Debug:    true,
	}

	return pgdb.Open(cfg)
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
