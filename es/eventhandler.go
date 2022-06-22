package es

import (
	"context"
	"reflect"
)

type EventHandler interface {
	Handle(ctx context.Context, evt Event, data interface{}) error
}

type eventHandler struct {
	name    string
	h       interface{}
	handles EventHandles
}

func (h *eventHandler) Handle(ctx context.Context, evt Event, data interface{}) error {
	t := reflect.TypeOf(data)
	r, ok := h.handles[t]
	if !ok {
		return ErrNotEventHandler
	}
	return r.Handle(h.h, ctx, evt, data)
}

func NewEventHandler(name string, h interface{}, handles EventHandles) EventHandler {
	return &eventHandler{
		name:    name,
		h:       h,
		handles: handles,
	}
}
