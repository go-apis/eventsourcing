package es

import "context"

type IsSaga interface {
	GetIsAsync() bool
}

type Saga interface {
	IsSaga

	Run(ctx context.Context, evt Event) ([]Command, error)
}

type BaseSaga struct {
	IsAsync bool
}

func (b *BaseSaga) GetIsAsync() bool {
	return b.IsAsync
}
