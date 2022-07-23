package streams

import (
	"context"
	"sync"

	"gorm.io/gorm"
)

type Manager interface {
	Listen(item *StreamItem) error
	Stop() error
}

type manager struct {
	mux sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	gormDb *gorm.DB

	senders map[string]Sender
}

func (m *manager) Listen(item *StreamItem) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	sender, ok := m.senders[item.ServiceName]
	if !ok {
		sender = NewSender(item.ServiceName)
		m.senders[item.ServiceName] = sender
	}

	// add the item
	if err := sender.Add(item); err != nil {
		return err
	}

	return nil
}

func (m *manager) Stop() error {
	m.cancel()
	return nil
}

func NewManager(gormDb *gorm.DB) (Manager, error) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	return &manager{
		ctx:     ctx,
		cancel:  cancel,
		gormDb:  gormDb,
		senders: make(map[string]Sender),
	}, nil
}
