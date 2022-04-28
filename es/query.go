package es

import "context"

type QueryObject interface {
	Aggregate
}

type Query[O QueryObject] interface {
	Find(ctx context.Context, id string) (O, error)
}

type query[O QueryObject] struct {
	store Store
	obj   O
}

func (q *query[O]) Find(ctx context.Context, id string) (O, error) {
	var obj O
	return obj, nil
}

func NewQuery[O QueryObject](store Store, obj O) Query[O] {
	return &query[O]{
		store: store,
		obj:   obj,
	}
}
