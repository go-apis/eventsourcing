package es

import "github.com/google/uuid"

type Snapshot struct {
	Namespace     string
	AggregateId   uuid.UUID
	AggregateType string
	Revision      string
	Aggregate     interface{}
}
