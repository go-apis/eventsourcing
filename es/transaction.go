package es

import "context"

type Tx interface {
	Commit(ctx context.Context) error
}

type Transaction interface {
	Tx

	DispatchAsync(ctx context.Context, cmds ...Command) error
	Dispatch(ctx context.Context, cmds ...Command) error
}

type transaction struct {
	*dispatcher
}

func (t *transaction) Commit(ctx context.Context) error {
	return nil
}

func NewTransaction(tx Tx) (Transaction, error) {
	return &transaction{}, nil
}
