package es

import (
	"github.com/google/uuid"
)

// EntityFunc for creating an entity
type EntityFunc func() (Entity, error)

type Entity interface {
	GetId() uuid.UUID
	GetNamespace() string
}
