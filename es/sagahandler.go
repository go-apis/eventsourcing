package es

import (
	"context"

	"go.opentelemetry.io/otel"
)

type sagaEventHandler struct {
	handles SagaHandles
	saga    IsSaga
}

func (b *sagaEventHandler) Handle(ctx context.Context, evt *Event) error {
	pctx, pspan := otel.Tracer("sagaEventHandler").Start(ctx, "Handle")
	defer pspan.End()

	unit, err := GetUnit(pctx)
	if err != nil {
		return err
	}

	cmds, err := b.handles.Handle(b.saga, pctx, evt)
	if err != nil {
		return err
	}

	if b.saga.GetIsAsync() {
		return unit.DispatchAsync(pctx, cmds...)
	}
	return unit.Dispatch(pctx, cmds...)
}

func NewSagaEventHandler(handles SagaHandles, saga IsSaga) EventHandler {
	return &sagaEventHandler{
		handles: handles,
		saga:    saga,
	}
}
