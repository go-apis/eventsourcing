package es

import "context"

type Saga interface {
	Run(ctx context.Context, evt Event) ([]Command, error)
}

type sagaEventHandler struct {
	handles SagaHandles
	saga    Saga
}

func (b *sagaEventHandler) Handle(ctx context.Context, evt Event) error {
	unit, err := GetUnit(ctx)
	if err != nil {
		return err
	}

	cmds, err := b.handles.Handle(b.saga, ctx, evt)
	if err != nil {
		return err
	}

	// dispatch.
	if err := unit.Dispatch(ctx, cmds...); err != nil {
		return err
	}

	return nil
}

func NewSagaEventHandler(handles SagaHandles, saga Saga) EventHandler {
	return &sagaEventHandler{
		handles: handles,
		saga:    saga,
	}
}
