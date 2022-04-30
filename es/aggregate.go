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

	version int
	events  []interface{}
}

func (a *BaseAggregate) GetEvents() []interface{} {
	return a.events
}

func (a *BaseAggregate) GetVersion() int {
	return a.version
}

func (a *BaseAggregate) IncrementVersion() {
	a.version++
}

func (a *BaseAggregate) Apply(ctx context.Context, data interface{}) error {
	a.events = append(a.events, data)
	return nil
}
