package es

import (
	"context"
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
	return nil
}

func NewEventHandler(name string, h interface{}, handles EventHandles) EventHandler {
	return &eventHandler{}
}
