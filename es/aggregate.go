package es

import (
	"context"
	"eventstore/es/event"
)

type Aggregate interface {
	Apply(ctx context.Context, data interface{}) error
}

type BaseAggregate struct {
	Namespace string `bun:",pk" json:"-"`
	Id        string `bun:",pk,type:uuid"`
	Version   int    `bun:"-"`

	events []*event.Event
}

func (a BaseAggregate) Apply(ctx context.Context, data interface{}) error {
	evt := &event.Event{
		Namespace:   a.Namespace,
		AggregateId: a.Id,
	}
	a.events = append(a.events, evt)
	return nil
}
