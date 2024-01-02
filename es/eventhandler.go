package es

import (
	"context"
)

type EventHandlers []EventHandler

func (h EventHandlers) Handle(ctx context.Context, evt *Event) error {
	for _, h := range h {
		if err := h.Handle(ctx, evt); err != nil {
			return err
		}
	}
	return nil
}

type EventHandler interface {
	Handle(ctx context.Context, evt *Event) error
}

type IsEventHandler interface {
	IsEventHandler()
}

type BaseEventHandler struct {
}

func (BaseEventHandler) IsEventHandler() {}

type eventHandler struct {
	h       interface{}
	handles EventHandlerHandles
}

func (h *eventHandler) Handle(ctx context.Context, evt *Event) error {
	return h.handles.Handle(h.h, ctx, evt)
}

func NewEventHandler(h interface{}, handles EventHandlerHandles) EventHandler {
	return &eventHandler{
		h:       h,
		handles: handles,
	}
}
