package es

import (
	"context"
	"encoding/json"
)

type Data interface {
	WithTx(ctx context.Context) (context.Context, Tx, error)
	GetTx(ctx context.Context) (Tx, error)
	LoadSnapshot(ctx context.Context, serviceName string, aggregateName string, namespace string, id string, out SourcedAggregate) error
	GetEventDatas(ctx context.Context, serviceName string, aggregateName string, namespace string, id string, fromVersion int) ([]json.RawMessage, error)
	SaveEvents(ctx context.Context, events []Event) error
}
