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
	MigrateDb(ctx context.Context) error
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

	SaveEvents(ctx context.Context, events []*Event) error
	SaveOutbox(ctx context.Context, rows []*OutboxEvent) error
	// ClaimOutbox locks and returns up to limit pending outbox rows in
	// insert order. It must run inside Begin: the claim is released by the
	// surrounding Commit/Rollback. Implementations return no rows when
	// another instance already holds the service's relay lock.
	ClaimOutbox(ctx context.Context, limit int) ([]*OutboxEvent, error)
	DeleteOutbox(ctx context.Context, ids []int64) error
	SaveEntity(ctx context.Context, aggregateName string, entity Entity) error
	SaveEntities(ctx context.Context, aggregateName string, entities []Entity) error
	DeleteEntity(ctx context.Context, aggregateName string, entity Entity) error
	Truncate(ctx context.Context, aggregateName string) error

	Get(ctx context.Context, aggregateName string, namespace string, id uuid.UUID, out interface{}) error
	One(ctx context.Context, aggregateName string, namespace string, filter Filter, out interface{}) error
	Find(ctx context.Context, aggregateName string, namespace string, filter Filter, out interface{}) error
	Count(ctx context.Context, aggregateName string, namespace string, filter Filter) (int, error)
	GroupedCount(ctx context.Context, aggregateName string, namespace string, filter Filter, groupBy string) ([]GroupCount, error)

	FindEvents(ctx context.Context, filter Filter) ([]*Event, error)
}

// GroupCount is a single (key, count) row from a grouped aggregation, i.e. the
// result of a SQL `GROUP BY <column>` with a `count(*)`.
type GroupCount struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}
