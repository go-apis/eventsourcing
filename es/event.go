package es

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type EventData struct {
	Version   int               `json:"version"`
	Type      string            `json:"type"`
	Data      json.RawMessage   `json:"data"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  datatypes.JSONMap `json:"metadata"`
}

type Event struct {
	ServiceName   string                 `json:"service_name"`
	Namespace     string                 `json:"namespace"`
	AggregateId   uuid.UUID              `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int                    `json:"version"`
	Type          string                 `json:"type"`
	Timestamp     time.Time              `json:"timestamp"`
	Data          interface{}            `json:"data"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// String implements the String method of the Event interface.
func (e Event) String() string {
	return fmt.Sprintf("%s@%d", e.Type, e.Version)
}
