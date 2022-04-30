package es

import (
	"context"
	"reflect"
)

type Query[O any] interface {
	Get(ctx context.Context, id string) (*O, error)
}

type query[O any] struct {
	store Store
	name  string
	obj   O
}

func (q *query[O]) Get(ctx context.Context, id string) (*O, error) {
	var obj O
	if err := q.store.Load(ctx, id, q.name, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func NewQuery[O any](store Store) Query[O] {
	var obj O
	t := reflect.TypeOf(obj)

	return &query[O]{
		store: store,
		name:  t.String(),
		obj:   obj,
	}
}
