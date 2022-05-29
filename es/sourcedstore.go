package es

import "context"

type SourcedStore interface {
	Load(ctx context.Context, id string, out SourcedAggregate) error
	Save(ctx context.Context, id string, val SourcedAggregate) error
}
