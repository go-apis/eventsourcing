package transactions

import (
	"time"

	"github.com/contextcloud/eventstore/server/pb/logger"
)

type Transaction struct {
	id               string
	creationTime     time.Time
	lastActivityTime time.Time
	log              logger.Logger
}

func NewTransaction(id string, log logger.Logger) *Transaction {
	now := time.Now()

	return &Transaction{
		log:              log,
		id:               id,
		creationTime:     now,
		lastActivityTime: now,
	}
}
