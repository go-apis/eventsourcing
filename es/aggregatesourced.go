package es

import (
	"context"
)

type AggregateSourced interface {
	Entity

	GetEvents() []interface{}
	GetVersion() int
	IncrementVersion()
}

type BaseAggregateSourced struct {
	BaseAggregate
	Version int `json:"version"`

	events []interface{}
}

func (a *BaseAggregateSourced) GetEvents() []interface{} {
	return a.events
}

func (a *BaseAggregateSourced) GetVersion() int {
	return a.Version
}

func (a *BaseAggregateSourced) IncrementVersion() {
	a.Version++
}

func (a *BaseAggregateSourced) Apply(ctx context.Context, data interface{}) error {
	a.events = append(a.events, data)
	return nil
}
