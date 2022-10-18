package es

import "github.com/google/uuid"

type AggregateParent struct {
	Id   uuid.UUID `json:"id"`
	Type string    `json:"type"`
}
