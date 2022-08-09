package es

import "context"

type IsSaga interface {
	IsSaga()
}

type Saga interface {
	IsSaga

	Run(ctx context.Context, evt Event) ([]Command, error)
}

type BaseSaga struct {
}

func (BaseSaga) IsSaga() {}
