package es

import (
	"context"
)

type EventHandlers []EventHandler

func (hs EventHandlers) Handle(ctx context.Context, evt *Event) error {
	for _, h := range hs {
		if err := h.HandleEvent(ctx, evt); err != nil {
			return err
		}
	}
	return nil
}

type EventHandler interface {
	HandleEvent(ctx context.Context, evt *Event) error
}

type GroupEventHandler interface {
	HandleGroupEvent(ctx context.Context, group string, evt *Event) error
}

type GroupMessageHandler interface {
	HandleGroupMessage(ctx context.Context, group string, msg []byte) error
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

func (h *eventHandler) HandleEvent(ctx context.Context, evt *Event) error {
	return h.handles.Handle(h.h, ctx, evt)
}

func NewEventHandler(h interface{}, handles EventHandlerHandles) EventHandler {
	return &eventHandler{
		h:       h,
		handles: handles,
	}
}
