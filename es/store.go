package es

import (
	"context"
	"fmt"

	"github.com/go-apis/eventsourcing/es/utils"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

type Store[T Entity] interface {
	Load(ctx context.Context, id uuid.UUID) (T, error)
	Save(ctx context.Context, entities ...T) error
}

type store[T Entity] struct {
	name string
}

func (q *store[T]) Load(ctx context.Context, id uuid.UUID) (T, error) {
	pctx, pspan := otel.Tracer("Query").Start(ctx, "Load")
	defer pspan.End()

	var item T
	unit, err := GetUnit(pctx)
	if err != nil {
		return item, err
	}

	out, err := unit.Load(pctx, q.name, id)
	if err != nil {
		return item, err
	}

	result, ok := out.(T)
	if !ok {
		return item, fmt.Errorf("unexpected type: %T", out)
	}
	return result, nil
}

func (q *store[T]) Save(ctx context.Context, entities ...T) error {
	pctx, pspan := otel.Tracer("Query").Start(ctx, "Save")
	defer pspan.End()

	unit, err := GetUnit(pctx)
	if err != nil {
		return err
	}

	for _, entity := range entities {
		if err := unit.Save(pctx, q.name, entity); err != nil {
			return err
		}
	}
	return nil
}

func NewStore[T Entity]() Store[T] {
	var item T
	typeOf := utils.GetElemType(item)

	return &store[T]{
		name: typeOf.Name(),
	}
}
