package es

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Actor struct {
	Id   uuid.UUID `json:"id"`
	Type string    `json:"type"`
}

// Event that has been persisted to the event store.
type Event struct {
	Service       string                 `json:"service"`
	Namespace     string                 `json:"namespace"`
	AggregateId   uuid.UUID              `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int                    `json:"version"`
	Type          string                 `json:"type"`
	By            *Actor                 `json:"by"`
	Timestamp     time.Time              `json:"timestamp"`
	Data          interface{}            `json:"data"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// String implements the String method of the Event interface.
func (e Event) String() string {
	return fmt.Sprintf("%s@%d", e.Type, e.Version)
}
