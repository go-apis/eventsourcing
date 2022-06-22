package es

import "context"

type Saga interface {
	Handle(ctx context.Context, cmd Command) error
	GetCommands() []Command
}

type SagaFactory func() (Saga, error)

type BaseSaga struct {
	commands []Command
}

func (a *BaseSaga) GetCommands() []Command {
	return a.commands
}

func (a *BaseSaga) Handle(ctx context.Context, cmd Command) error {
	a.commands = append(a.commands, cmd)
	return nil
}

type sagaEventHandler struct {
	handle  *EventHandle
	factory SagaFactory
}

func (b *sagaEventHandler) Handle(ctx context.Context, evt Event, data interface{}) error {
	unit := UnitFromContext(ctx)

	saga, err := b.factory()
	if err != nil {
		return err
	}

	if err := b.handle.Handle(saga, ctx, evt, data); err != nil {
		return err
	}

	// dispatch.
	cmds := saga.GetCommands()
	if err := unit.Dispatch(ctx, cmds...); err != nil {
		return err
	}

	return nil
}

func NewSagaEventHandler(handle *EventHandle, factory SagaFactory) EventHandler {
	return &sagaEventHandler{
		handle:  handle,
		factory: factory,
	}
}
