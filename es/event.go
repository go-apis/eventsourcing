package es

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Metadata is a simple map to store event's metadata
type Metadata = map[string]interface{}

type Event struct {
	ServiceName   string      `json:"service_name"`
	Namespace     string      `json:"namespace"`
	AggregateId   uuid.UUID   `json:"aggregate_id"`
	AggregateType string      `json:"aggregate_type"`
	Version       int         `json:"version"`
	Type          string      `json:"type"`
	Timestamp     time.Time   `json:"timestamp"`
	Data          interface{} `json:"data"`
	Metadata      Metadata    `json:"metadata"`
}

// String implements the String method of the Event interface.
func (e Event) String() string {
	return fmt.Sprintf("%s@%d", e.Type, e.Version)
}
