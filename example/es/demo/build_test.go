package demo

import (
	"context"
	"eventstore/example/es"
	"testing"
)

func Test_Build(t *testing.T) {
	h, err := es.NewCommandHandler(&TestAggregate{})
	if err != nil {
		t.Error(err)
		return
	}

	if h == nil {
		t.Error("handler is nil")
		return
	}

	cmd := &CreateTest{
		BaseCommand: es.BaseCommand{
			AggregateId: "test",
		},
		Name: "test",
	}
	ctx := context.Background()

	if err := h.Handle(ctx, cmd); err != nil && err.Error() != "Whoops" {
		t.Error(err)
		return
	}
}
