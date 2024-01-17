package es

import (
	"context"

	"github.com/google/uuid"
)

type SnapshotSearch struct {
	Namespace     string
	AggregateType string
	AggregateId   uuid.UUID
	Revision      string
}

type ConnFactory func(ctx context.Context, cfg *ProviderConfig, reg Registry) (Conn, error)

type Conn interface {
	NewData(ctx context.Context) (Data, error)
	Close(ctx context.Context) error
}

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Lock interface {
	Unlock(ctx context.Context) error
}

type Data interface {
	Begin(ctx context.Context) (Tx, error)
	Lock(ctx context.Context) (Lock, error)

	LoadSnapshot(ctx context.Context, search SnapshotSearch, out AggregateSourced) error
	SaveSnapshot(ctx context.Context, snapshot *Snapshot) error

	SavePersistedCommand(ctx context.Context, cmd *PersistedCommand) error
	DeletePersistedCommand(ctx context.Context, cmd *PersistedCommand) error
	FindPersistedCommands(ctx context.Context, filter Filter) ([]*PersistedCommand, error)
	NewScheduledCommandNotifier(ctx context.Context) (*ScheduledCommandNotifier, error)

	SaveEvents(ctx context.Context, events []*Event) error
	SaveEntity(ctx context.Context, aggregateName string, entity Entity) error
	DeleteEntity(ctx context.Context, aggregateName string, entity Entity) error
	Truncate(ctx context.Context, aggregateName string) error

	Get(ctx context.Context, aggregateName string, namespace string, id uuid.UUID, out interface{}) error
	Find(ctx context.Context, aggregateName string, namespace string, filter Filter, out interface{}) error
	Count(ctx context.Context, aggregateName string, namespace string, filter Filter) (int, error)

	FindEvents(ctx context.Context, filter Filter) ([]*Event, error)
}
