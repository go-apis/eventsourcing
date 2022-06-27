package es

import (
	"context"
	"fmt"
	"math"
	"reflect"

	"github.com/contextcloud/eventstore/es/filters"
)

type Pagination[T any] struct {
	Limit      int64
	Page       int64
	TotalItems int64
	TotalPages int64
	Items      []T
}

type Query[T any] interface {
	Load(ctx context.Context, id string) (*T, error)
	Find(ctx context.Context, filter filters.Filter) ([]T, error)
	Count(ctx context.Context, filter filters.Filter) (int64, error)
	Pagination(ctx context.Context, filter filters.Filter) (*Pagination[T], error)
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

func (q *query[T]) Count(ctx context.Context, filter filters.Filter) (int64, error) {
	return q.unit.Count(ctx, q.aggregateName, filter)
}

func (q *query[T]) Pagination(ctx context.Context, filter filters.Filter) (*Pagination[T], error) {
	if filter.Limit == nil {
		return nil, fmt.Errorf("Limit required for pagination")
	}
	if filter.Offset == nil {
		return nil, fmt.Errorf("Offset required for pagination")
	}

	totalItems, err := q.unit.Count(ctx, q.aggregateName, filter)
	if err != nil {
		return nil, err
	}

	var items []T
	if err := q.unit.Find(ctx, q.aggregateName, filter, &items); err != nil {
		return nil, err
	}

	totalPages := int64(math.Ceil(float64(totalItems) / float64(*filter.Limit)))
	return &Pagination[T]{
		Limit:      int64(*filter.Limit),
		Page:       int64(*filter.Offset) + 1,
		TotalItems: totalItems,
		TotalPages: totalPages,
		Items:      items,
	}, nil
}

func NewQuery[T any](unit Unit) Query[T] {
	var item T
	typeOf := reflect.TypeOf(item)

	return &query[T]{
		unit:          unit,
		aggregateName: typeOf.String(),
	}
}
