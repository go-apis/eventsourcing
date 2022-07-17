package transactions

import (
	"sync"
	"time"

	"github.com/contextcloud/eventstore/server/pb/logger"
	"gorm.io/gorm"
)

type Transaction struct {
	mux              sync.RWMutex
	gormDb           *gorm.DB
	id               string
	creationTime     time.Time
	lastActivityTime time.Time
	log              logger.Logger
}

func (tx *Transaction) GetDb() *gorm.DB {
	tx.mux.Lock()
	defer tx.mux.Unlock()

	return tx.gormDb
}

func (tx *Transaction) GetID() string {
	tx.mux.Lock()
	defer tx.mux.Unlock()

	return tx.id
}

func (tx *Transaction) GetLastActivityTime() time.Time {
	tx.mux.RLock()
	defer tx.mux.RUnlock()

	return tx.lastActivityTime
}

func (tx *Transaction) GetCreationTime() time.Time {
	tx.mux.RLock()
	defer tx.mux.RUnlock()

	return tx.creationTime
}

func (tx *Transaction) SetLastActivityTime(t time.Time) {
	tx.mux.Lock()
	defer tx.mux.Unlock()

	tx.lastActivityTime = t
}

func (tx *Transaction) Rollback() error {
	tx.mux.Lock()
	defer tx.mux.Unlock()

	q := tx.gormDb.Rollback()
	return q.Error
}

func (tx *Transaction) Commit() (int, error) {
	tx.mux.Lock()
	defer tx.mux.Unlock()

	q := tx.gormDb.Commit()
	return int(q.RowsAffected), q.Error
}

func NewTransaction(id string, gormDb *gorm.DB, log logger.Logger) *Transaction {
	now := time.Now()

	return &Transaction{
		log:              log,
		id:               id,
		gormDb:           gormDb,
		creationTime:     now,
		lastActivityTime: now,
	}
}
