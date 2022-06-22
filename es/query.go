package es

import (
	"context"
	"reflect"

	"github.com/contextcloud/eventstore/es/filters"
)

type Query[T any] interface {
	Load(ctx context.Context, id string) (*T, error)
	Find(ctx context.Context, filter filters.Filter) ([]T, error)
}

type query[T any] struct {
	unit          Unit
	aggregateName string
}

func (q *query[T]) Load(ctx context.Context, id string) (*T, error) {
	var item T
	if err := q.unit.Load(ctx, id, q.aggregateName, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (q *query[T]) Find(ctx context.Context, filter filters.Filter) ([]T, error) {
	var items []T
	if err := q.unit.Find(ctx, q.aggregateName, filter, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func NewQuery[T any](unit Unit) Query[T] {
	var item T
	typeOf := reflect.TypeOf(item)

	return &query[T]{
		unit:          unit,
		aggregateName: typeOf.String(),
	}
}
