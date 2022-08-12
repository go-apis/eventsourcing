package es

import (
	"context"
)

type EventHandler interface {
	Handle(ctx context.Context, evt Event) error
}

type eventHandler struct {
	name    string
	h       interface{}
	handles EventHandles
}

func (h *eventHandler) Handle(ctx context.Context, evt Event) error {
	return h.handles.Handle(h.h, ctx, evt)
}

func NewEventHandler(name string, h interface{}, handles EventHandles) EventHandler {
	return &eventHandler{
		name:    name,
		h:       h,
		handles: handles,
	}
}
