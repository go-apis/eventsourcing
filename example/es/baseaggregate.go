package es

import "context"

type Aggregate interface {
	Apply(ctx context.Context, data interface{}) error
}

type BaseAggregate struct {
}

func (a BaseAggregate) Apply(ctx context.Context, data interface{}) error {
	return nil
}
