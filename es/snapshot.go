package es

import "github.com/google/uuid"

type Snapshot struct {
	ServiceName   string
	Namespace     string
	AggregateId   uuid.UUID
	AggregateType string
	Revision      string
	Aggregate     interface{}
}
