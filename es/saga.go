package es

import "context"

type Saga interface {
	Run(ctx context.Context, cmd Command) ([]Command, error)
}

type sagaEventHandler struct {
	handles EventHandles
	saga    Saga
}

func (b *sagaEventHandler) Handle(ctx context.Context, evt Event, data interface{}) error {
	// unit, err := GetUnit(ctx)
	// if err != nil {
	// 	return err
	// }

	// if err := b.handles.Handle(b.s, ctx, evt, data); err != nil {
	// 	return err
	// }

	// // dispatch.
	// cmds := sag()
	// if err := unit.Dispatch(ctx, cmds...); err != nil {
	// 	return err
	// }

	return nil
}

func NewSagaEventHandler(handles EventHandles, saga Saga) EventHandler {
	return &sagaEventHandler{
		handles: handles,
		saga:    saga,
	}
}
