package transactions

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/contextcloud/eventstore/server/pb/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type manager struct {
	mux sync.RWMutex

	running      bool
	options      Options
	gormDb       *gorm.DB
	transactions map[string]*Transaction
	ticker       *time.Ticker
	done         chan bool
	logger       logger.Logger
}

type Manager interface {
	NewTransaction(ctx context.Context) (*Transaction, error)
	TransactionExists(id string) bool
	GetTransaction(id string) (*Transaction, error)
	TransactionCount() int

	DeleteTransaction(id string) error
	CommitTransaction(id string) (int, error)
	UpdateSessionActivityTime(id string)

	IsRunning() bool
	Start() error
	Stop() error
}

func NewManager(options *Options, gormDb *gorm.DB) (*manager, error) {
	if options == nil {
		return nil, ErrInvalidOptionsProvided
	}

	guard := &manager{
		gormDb:       gormDb,
		transactions: make(map[string]*Transaction),
		ticker:       time.NewTicker(options.SessionGuardCheckInterval),
		done:         make(chan bool),
		logger:       logger.NewSimpleLogger("transaction guard", os.Stdout),
		options:      *options,
	}

	return guard, nil
}

func (sm *manager) IsRunning() bool {
	sm.mux.RLock()
	defer sm.mux.RUnlock()

	return sm.running
}

func (sm *manager) NewTransaction(ctx context.Context) (*Transaction, error) {
	sm.mux.Lock()
	defer sm.mux.Unlock()

	if len(sm.transactions) >= sm.options.MaxSessions {
		sm.logger.Warningf("max sessions reached")
		return nil, ErrMaxTransactionsReached
	}

	id := uuid.NewString()

	gormDb := sm.gormDb.Begin()
	if gormDb.Error != nil {
		return nil, gormDb.Error
	}

	sm.transactions[id] = NewTransaction(id, gormDb, sm.logger)
	sm.logger.Debugf("created session %s", id)

	return sm.transactions[id], nil
}

func (sm *manager) TransactionExists(id string) bool {
	sm.mux.RLock()
	defer sm.mux.RUnlock()

	_, ok := sm.transactions[id]
	return ok
}

func (sm *manager) GetTransaction(id string) (*Transaction, error) {
	sm.mux.RLock()
	defer sm.mux.RUnlock()

	transaction, ok := sm.transactions[id]
	if !ok {
		return nil, ErrTransactionNotFound
	}

	return transaction, nil
}

func (sm *manager) DeleteTransaction(id string) error {
	sm.mux.Lock()
	defer sm.mux.Unlock()

	transaction, ok := sm.transactions[id]
	if !ok {
		return ErrTransactionNotFound
	}

	err := transaction.Rollback()
	delete(sm.transactions, id)
	if err != nil {
		return err
	}
	return nil
}

func (sm *manager) CommitTransaction(id string) (int, error) {
	sm.mux.Lock()
	defer sm.mux.Unlock()

	transaction, ok := sm.transactions[id]
	if !ok {
		return 0, ErrTransactionNotFound
	}
	defer transaction.Rollback()

	out, err := transaction.Commit()
	delete(sm.transactions, id)
	if err != nil {
		return out, err
	}
	return out, nil
}

func (sm *manager) UpdateSessionActivityTime(id string) {
	sm.mux.Lock()
	defer sm.mux.Unlock()

	if sess, ok := sm.transactions[id]; ok {
		now := time.Now()
		sess.SetLastActivityTime(now)
		sm.logger.Debugf("updated last activity time for %s at %s", id, now.Format(time.UnixDate))
	}
}

func (sm *manager) TransactionCount() int {
	sm.mux.RLock()
	defer sm.mux.RUnlock()

	return len(sm.transactions)
}

func (sm *manager) Start() error {
	sm.mux.Lock()
	defer sm.mux.Unlock()

	if sm.running {
		return ErrGuardAlreadyRunning
	}
	sm.running = true

	go func() {
		for {
			select {
			case <-sm.done:
				return
			case <-sm.ticker.C:
				sm.expireTransactions(time.Now())
			}
		}
	}()

	return nil
}
func (sm *manager) Stop() error {
	sm.mux.Lock()
	defer sm.mux.Unlock()

	if !sm.running {
		return ErrGuardNotRunning
	}
	sm.running = false
	sm.ticker.Stop()

	// Wait for the guard to finish any pending cancellation work
	// this must be done with unlocked mutex since
	// mutex expiration may try to lock the mutex
	sm.mux.Unlock()
	sm.done <- true
	sm.mux.Lock()

	// Delete all
	for id, transaction := range sm.transactions {
		transaction.Rollback()
		delete(sm.transactions, id)
	}

	sm.logger.Debugf("shutdown")
	return nil
}

func (sm *manager) expireTransactions(now time.Time) (transactionCount, inactiveTransactionCount, deletedTransactionCount int, err error) {
	sm.mux.Lock()
	defer sm.mux.Unlock()

	if !sm.running {
		return 0, 0, 0, ErrGuardNotRunning
	}

	inactiveTransactionCount = 0
	deletedTransactionCount = 0
	sm.logger.Debugf("checking at %s", now.Format(time.UnixDate))

	for id, transaction := range sm.transactions {
		createdAt := transaction.GetCreationTime()
		lastActivity := transaction.GetLastActivityTime()

		if now.Sub(lastActivity) > sm.options.MaxSessionInactivityTime {
			inactiveTransactionCount++
			continue
		}

		if now.Sub(createdAt) > sm.options.MaxSessionAgeTime {
			sm.logger.Debugf("removing session %s - exceeded MaxSessionAgeTime", id)
			transaction.Rollback()
			delete(sm.transactions, id)
			deletedTransactionCount++
		}

		if now.Sub(lastActivity) > sm.options.Timeout {
			sm.logger.Debugf("removing session %s - exceeded Timeout", id)
			transaction.Rollback()
			delete(sm.transactions, id)
			deletedTransactionCount++
		}
	}

	sm.logger.Debugf("Open sessions count: %d\n", len(sm.transactions))
	sm.logger.Debugf("Inactive sessions count: %d\n", inactiveTransactionCount)
	sm.logger.Debugf("Deleted sessions count: %d\n", deletedTransactionCount)

	return len(sm.transactions), inactiveTransactionCount, deletedTransactionCount, nil
}
