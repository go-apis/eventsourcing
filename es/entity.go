package es

import "github.com/google/uuid"

type Entity struct {
	ServiceName   string      `json:"service_name"`
	Namespace     string      `json:"namespace"`
	AggregateId   uuid.UUID   `json:"aggregate_id"`
	AggregateType string      `json:"aggregate_type"`
	Data          interface{} `json:"data"`
}
