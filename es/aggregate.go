package es

import (
	"context"

	"github.com/google/uuid"
)

type Aggregate interface {
	Apply(ctx context.Context, data interface{}) error
}

type SetId interface {
	SetId(id uuid.UUID, namespace string)
}

type SetVersion interface {
	SetVersion(version int)
}

type BaseAggregate struct {
	Id        uuid.UUID `json:"id" format:"uuid" required:"true"`
	Namespace string    `json:"namespace" required:"true"`
}

// GetId of the aggregate
func (a *BaseAggregate) GetId() uuid.UUID {
	return a.Id
}

// SetId of the aggregate
func (a *BaseAggregate) SetId(id uuid.UUID, namespace string) {
	a.Id = id
	a.Namespace = namespace
}
