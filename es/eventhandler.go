package es

import (
	"context"

	"go.opentelemetry.io/otel"
)

type EventHandler interface {
	Handle(ctx context.Context, evt *Event) error
}

type eventHandler struct {
	name    string
	h       interface{}
	handles EventHandles
}

func (h *eventHandler) Handle(ctx context.Context, evt *Event) error {
	pctx, pspan := otel.Tracer("EventHandler").Start(ctx, "Handle")
	defer pspan.End()

	return h.handles.Handle(h.h, pctx, evt)
}

func NewEventHandler(name string, h interface{}, handles EventHandles) EventHandler {
	return &eventHandler{
		name:    name,
		h:       h,
		handles: handles,
	}
}
