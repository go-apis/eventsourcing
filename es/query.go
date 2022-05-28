package es

import (
	"context"
	"reflect"
)

type Query[O any] interface {
	Get(ctx context.Context, id string) (*O, error)
}

type query[O any] struct {
	store Data
	name  string
	obj   O
}

func (q *query[O]) Get(ctx context.Context, id string) (*O, error) {
	tx, err := q.store.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	var obj O
	if err := tx.Load(ctx, id, q.name, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func NewQuery[O any](store Data) Query[O] {
	var obj O
	t := reflect.TypeOf(obj)

	return &query[O]{
		store: store,
		name:  t.String(),
		obj:   obj,
	}
}
