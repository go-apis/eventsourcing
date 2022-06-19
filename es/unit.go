package es

import "context"

type Unit interface {
	DispatchAsync(ctx context.Context, cmds ...Command) error
	Dispatch(ctx context.Context, cmds ...Command) error
	Commit(ctx context.Context) error
}

type unit struct {
	tx Tx
	*dispatcher
}

func (u *unit) Commit(ctx context.Context) error {
	return nil
}

func NewUnit(tx Tx) (Unit, error) {
	return &unit{
		tx: tx,
	}, nil
}
