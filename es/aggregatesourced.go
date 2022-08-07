package es

import (
	"context"

	"github.com/google/uuid"
)

type AggregateSourced interface {
	GetEvents() []interface{}
	GetVersion() int
	IncrementVersion()
}

type BaseAggregateSourced struct {
	Id        uuid.UUID `json:"id"`
	Namespace string    `json:"namespace"`

	version int
	events  []interface{}
}

func (a *BaseAggregateSourced) GetId() uuid.UUID {
	return a.Id
}

func (a *BaseAggregateSourced) SetId(id uuid.UUID, namespace string) {
	a.Id = id
	a.Namespace = namespace
}

func (a *BaseAggregateSourced) GetEvents() []interface{} {
	return a.events
}

func (a *BaseAggregateSourced) GetVersion() int {
	return a.version
}

func (a *BaseAggregateSourced) IncrementVersion() {
	a.version++
}

func (a *BaseAggregateSourced) Apply(ctx context.Context, data interface{}) error {
	a.events = append(a.events, data)
	return nil
}
