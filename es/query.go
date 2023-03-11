package es

import (
	"context"
	"fmt"
	"math"

	"github.com/contextcloud/eventstore/es/filters"
	"github.com/contextcloud/eventstore/es/utils"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

type Pagination[T any] struct {
	Limit      int `json:"limit"`
	Page       int `json:"page"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
	Items      []T `json:"items"`
}

type Query[T Entity] interface {
	Get(ctx context.Context, id uuid.UUID) (T, error)
	Find(ctx context.Context, filter filters.Filter) ([]T, error)
	Count(ctx context.Context, filter filters.Filter) (int, error)
	Pagination(ctx context.Context, filter filters.Filter) (*Pagination[T], error)
}

type query[T Entity] struct {
	options *QueryOptions
	name    string
}

func (q *query[T]) Get(ctx context.Context, id uuid.UUID) (T, error) {
	pctx, pspan := otel.Tracer("Query").Start(ctx, "Get")
	defer pspan.End()

	namespace := ""
	if q.options.UseNamespace {
		namespace = NamespaceFromContext(ctx)
	}

	var item T
	unit, err := GetUnit(pctx)
	if err != nil {
		return item, err
	}

	if err := unit.GetData().Get(pctx, q.name, namespace, id, &item); err != nil {
		return item, err
	}
	return item, nil
}

func (q *query[T]) Find(ctx context.Context, filter filters.Filter) ([]T, error) {
	pctx, pspan := otel.Tracer("Query").Start(ctx, "Find")
	defer pspan.End()

	namespace := ""
	if q.options.UseNamespace {
		namespace = NamespaceFromContext(ctx)
	}

	unit, err := GetUnit(pctx)
	if err != nil {
		return nil, err
	}

	var items []T
	if err := unit.GetData().Find(pctx, q.name, namespace, filter, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (q *query[T]) Count(ctx context.Context, filter filters.Filter) (int, error) {
	pctx, pspan := otel.Tracer("Query").Start(ctx, "Count")
	defer pspan.End()

	namespace := ""
	if q.options.UseNamespace {
		namespace = NamespaceFromContext(ctx)
	}

	unit, err := GetUnit(pctx)
	if err != nil {
		return 0, err
	}

	return unit.GetData().Count(pctx, q.name, namespace, filter)
}

func (q *query[T]) Pagination(ctx context.Context, filter filters.Filter) (*Pagination[T], error) {
	pctx, pspan := otel.Tracer("Query").Start(ctx, "Pagination")
	defer pspan.End()

	namespace := ""
	if q.options.UseNamespace {
		namespace = NamespaceFromContext(ctx)
	}

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

	totalItems, err := unit.GetData().Count(pctx, q.name, namespace, filter)
	if err != nil {
		return nil, err
	}

	var items []T
	if err := unit.GetData().Find(pctx, namespace, q.name, filter, &items); err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(*filter.Limit)))
	return &Pagination[T]{
		Limit:      *filter.Limit,
		Page:       *filter.Offset + 1,
		TotalItems: totalItems,
		TotalPages: totalPages,
		Items:      items,
	}, nil
}

func NewQuery[T Entity](options ...QueryOption) Query[T] {
	var item T
	typeOf := utils.GetElemType(item)

	o := DefaultQueryOptions()
	for _, option := range options {
		option(o)
	}

	return &query[T]{
		options: o,
		name:    typeOf.Name(),
	}
}
