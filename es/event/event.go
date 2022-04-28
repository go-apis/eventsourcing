package event

import "time"

type Event struct {
	Namespace     string      `json:"namespace"`
	AggregateId   string      `json:"aggregate_id"`
	AggregateType string      `json:"aggregate_type"`
	Version       int         `json:"version"`
	Type          string      `json:"type"`
	Timestamp     time.Time   `json:"timestamp"`
	Data          interface{} `json:"data"`
	// Metadata      Metadata    `json:"metadata"`
}
