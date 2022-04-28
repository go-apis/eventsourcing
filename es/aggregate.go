package es

import (
	"context"
)

type Aggregate interface {
	Apply(ctx context.Context, data interface{}) error
}

type SourcedAggregate interface {
	GetEvents() []interface{}
	GetVersion() int
	IncrementVersion()
}

type BaseAggregate struct {
	Namespace string `bun:",pk" json:"-"`
	Id        string `bun:",pk,type:uuid"`
	Version   int    `bun:"-"`

	events []interface{}
}

// GetVersion returns the version of the aggregate.
func (a *BaseAggregate) GetVersion() int {
	return a.Version
}

// IncrementVersion ads 1 to the current version
func (a *BaseAggregate) IncrementVersion() {
	a.Version++
}

func (a *BaseAggregate) Apply(ctx context.Context, data interface{}) error {
	a.events = append(a.events, data)
	return nil
}
