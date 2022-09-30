package es

import (
	"context"

	"github.com/contextcloud/eventstore/es/filters"
	"github.com/google/uuid"
)

type InitializeOptions struct {
	ServiceName   string
	Revision      string
	EventConfigs  []*EventConfig
	EntityConfigs []*EntityConfig
}

type EventSearch struct {
	ServiceName   string
	Namespace     string
	AggregateType string
	AggregateId   uuid.UUID
	FromVersion   int
}

type SnapshotSearch struct {
	ServiceName   string
	Namespace     string
	AggregateType string
	AggregateId   uuid.UUID
	Revision      string
}

type ConnFactory func(cfg DataConfig) (Conn, error)

type Conn interface {
	Initialize(ctx context.Context, opts InitializeOptions) error
	NewData(ctx context.Context) (Data, error)
	Close(ctx context.Context) error
}

type Tx interface {
	Commit(ctx context.Context) (int, error)
	Rollback(ctx context.Context) error
}

type Data interface {
	Begin(ctx context.Context) (Tx, error)

	LoadSnapshot(ctx context.Context, search SnapshotSearch, out AggregateSourced) error
	SaveSnapshot(ctx context.Context, snapshot *Snapshot) error

	GetEvents(ctx context.Context, mapper EventDataMapper, search EventSearch) ([]*Event, error)
	SaveEvents(ctx context.Context, events []*Event) error
	SaveEntity(ctx context.Context, serviceName string, aggregateName string, entity Entity) error

	Load(ctx context.Context, serviceName string, aggregateName string, namespace string, id uuid.UUID, out interface{}) error
	Find(ctx context.Context, serviceName string, aggregateName string, namespace string, filter filters.Filter, out interface{}) error
	Count(ctx context.Context, serviceName string, aggregateName string, namespace string, filter filters.Filter) (int, error)
}
