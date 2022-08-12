package es

import (
	"context"

	"github.com/contextcloud/eventstore/es/filters"
	"github.com/google/uuid"
)

type Conn interface {
	Initialize(ctx context.Context, serviceName string, opts ...EntityOptions) error
	NewData(ctx context.Context) (Data, error)
	Publish(ctx context.Context, evts ...Event) error
	Close(ctx context.Context) error
}

type Tx interface {
	Commit(ctx context.Context) (int, error)
	Rollback(ctx context.Context) error
}

type Data interface {
	Begin(ctx context.Context) (Tx, error)

	LoadSnapshot(ctx context.Context, serviceName string, aggregateName string, namespace string, revision string, id uuid.UUID, out AggregateSourced) error
	SaveSnapshot(ctx context.Context, serviceName string, aggregateName string, namespace string, revision string, id uuid.UUID, out AggregateSourced) error

	GetEventDatas(ctx context.Context, serviceName string, aggregateName string, namespace string, id uuid.UUID, fromVersion int) ([]*EventData, error)
	SaveEventDatas(ctx context.Context, serviceName string, aggregateName string, namespace string, id uuid.UUID, datas []*EventData) error
	SaveEntity(ctx context.Context, serviceName string, aggregateName string, entity Entity) error

	Load(ctx context.Context, serviceName string, aggregateName string, namespace string, id uuid.UUID, out interface{}) error
	Find(ctx context.Context, serviceName string, aggregateName string, namespace string, filter filters.Filter, out interface{}) error
	Count(ctx context.Context, serviceName string, aggregateName string, namespace string, filter filters.Filter) (int, error)
}
