package es

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

func toPage(limit, offset int) int {
	return int(math.Floor(float64(offset)/float64(limit))) + 1
}

type Pagination[T any] struct {
	Limit      int   `json:"limit"`
	Page       int   `json:"page"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	Items      []T   `json:"items"`
}

type Query[T Entity] interface {
	Get(ctx context.Context, id uuid.UUID) (T, error)
	Find(ctx context.Context, filter Filter) ([]T, error)
	Count(ctx context.Context, filter Filter) (int, error)
	Pagination(ctx context.Context, filter Filter) (*Pagination[T], error)
}

type query[T Entity] struct {
	options *QueryOptions
	name    string
}

func (q *query[T]) getNamespace(ctx context.Context) string {
	if len(q.options.Namespace) > 0 {
		return q.options.Namespace
	}
	if q.options.UseNamespace {
		return GetNamespace(ctx)
	}
	return ""
}

func (q *query[T]) Get(ctx context.Context, id uuid.UUID) (T, error) {
	pctx, pspan := otel.Tracer("Query").Start(ctx, "Get")
	defer pspan.End()

	var item T
	unit, err := GetUnit(pctx)
	if err != nil {
		return item, err
	}

	namespace := q.getNamespace(pctx)
	if err := unit.Get(pctx, q.name, namespace, id, &item); err != nil {
		return item, err
	}
	return item, nil
}

func (q *query[T]) Find(ctx context.Context, filter Filter) ([]T, error) {
	pctx, pspan := otel.Tracer("Query").Start(ctx, "Find")
	defer pspan.End()

	unit, err := GetUnit(pctx)
	if err != nil {
		return nil, err
	}

	namespace := q.getNamespace(pctx)

	var items []T
	if err := unit.Find(pctx, q.name, namespace, filter, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (q *query[T]) Count(ctx context.Context, filter Filter) (int, error) {
	pctx, pspan := otel.Tracer("Query").Start(ctx, "Count")
	defer pspan.End()

	unit, err := GetUnit(pctx)
	if err != nil {
		return 0, err
	}

	namespace := q.getNamespace(pctx)
	return unit.Count(pctx, q.name, namespace, filter)
}

func (q *query[T]) Pagination(ctx context.Context, filter Filter) (*Pagination[T], error) {
	pctx, pspan := otel.Tracer("Query").Start(ctx, "Pagination")
	defer pspan.End()

	if filter.Limit == nil {
		return nil, fmt.Errorf("Limit required for pagination")
	}
	if filter.Offset == nil {
		return nil, fmt.Errorf("Offset required for pagination")
	}

	unit, err := GetUnit(pctx)
	if err != nil {
		return nil, err
	}

	namespace := q.getNamespace(pctx)
	totalItems, err := unit.Count(pctx, q.name, namespace, filter)
	if err != nil {
		return nil, err
	}

	var items []T
	if err := unit.Find(pctx, q.name, namespace, filter, &items); err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(*filter.Limit)))
	return &Pagination[T]{
		Limit:      *filter.Limit,
		Page:       toPage(*filter.Limit, *filter.Offset),
		TotalItems: int64(totalItems),
		TotalPages: totalPages,
		Items:      items,
	}, nil
}

func NewQuery[T Entity](options ...QueryOption) Query[T] {
	var entity T
	opts := NewEntityOptions(entity)
	entityConfig, err := NewEntityConfig(opts)
	if err != nil {
		panic(err)
	}

	o := DefaultQueryOptions()
	for _, option := range options {
		option(o)
	}

	return &query[T]{
		options: o,
		name:    entityConfig.Name,
	}
}
