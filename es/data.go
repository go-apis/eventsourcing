package es

import "context"

type Data interface {
	WithTx(ctx context.Context) (context.Context, Tx, error)
	LoadSnapshot(ctx context.Context, serviceName string, aggregateName string, namespace string, id string, out SourcedAggregate) error
	GetEvents(ctx context.Context, serviceName string, aggregateName string, namespace string, id string, fromVersion int) ([]Event, error)
	SaveEvents(ctx context.Context, events []Event) error
}
