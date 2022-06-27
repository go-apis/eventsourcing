package es

import (
	"context"
	"encoding/json"

	"github.com/contextcloud/eventstore/es/filters"
)

type Tx interface {
	Commit(ctx context.Context) error

	LoadSnapshot(ctx context.Context, serviceName string, aggregateName string, namespace string, id string, out SourcedAggregate) error
	GetEventDatas(ctx context.Context, serviceName string, aggregateName string, namespace string, id string, fromVersion int) ([]json.RawMessage, error)
	SaveEvents(ctx context.Context, events []Event) error
	SaveEntity(ctx context.Context, entity Entity) error

	Load(ctx context.Context, serviceName string, aggregateName string, namespace string, id string, out interface{}) error
	Find(ctx context.Context, serviceName string, aggregateName string, namespace string, filter filters.Filter, out interface{}) error
	Count(ctx context.Context, serviceName string, aggregateName string, namespace string, filter filters.Filter) (int64, error)
}
