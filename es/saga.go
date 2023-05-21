package es

import "context"

type IsSaga interface {
	Saga()
}

type Saga interface {
	IsSaga

	Run(ctx context.Context, evt Event) ([]Command, error)
}

type BaseSaga struct {
}

func (b *BaseSaga) Saga() {
}
