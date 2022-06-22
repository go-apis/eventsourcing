package es

import (
	"context"
)

type Aggregate interface {
	Apply(ctx context.Context, data interface{}) error
}

type SourcedAggregateFactory func() (SourcedAggregate, error)

type SourcedAggregate interface {
	GetEvents() []interface{}
	GetVersion() int
	IncrementVersion()
}

type SetAggregate interface {
	SetId(id string, namespace string)
}

type BaseAggregate struct {
	Id        string
	Namespace string

	version int
	events  []interface{}
}

func (a *BaseAggregate) SetId(id string, namespace string) {
	a.Id = id
	a.Namespace = namespace
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
