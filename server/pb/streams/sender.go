package streams

import (
	"sync"
	"time"
)

type Sender interface {
	Add(item *StreamItem) error
}

type sender struct {
	mux sync.RWMutex

	serviceName string
	ticker      *time.Ticker
}

func (s *sender) Add(item *StreamItem) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	return nil
}

func NewSender(serviceName string) Sender {
	return &sender{
		serviceName: serviceName,
		ticker:      &time.Ticker{},
	}
}
