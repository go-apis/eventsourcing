package eventhandlers

import (
	"context"
	"log"

	"github.com/go-apis/eventsourcing/es"
	"github.com/go-apis/eventsourcing/examples/users/data/events"
)

type demoHandler struct {
	es.BaseEventHandler `es:"group=random"`
}

func (h *demoHandler) Handle(ctx context.Context, event *es.Event, data *events.UserCreated) error {
	log.Printf("UserCreated: %v", data)
	return nil
}

func NewDemoHandler() es.IsEventHandler {
	return &demoHandler{}
}
