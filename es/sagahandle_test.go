package es

import (
	"context"
	"testing"
)

type demoEventData struct {
}

type demoSaga struct {
	BaseSaga
}

func (s *demoSaga) HandleEventData(ctx context.Context, evt *Event, data *demoEventData) ([]Command, error) {
	return nil, nil
}

func Test_SagaHandle(t *testing.T) {
	saga := &demoSaga{}
	handles := NewSagaHandles(saga)

	if len(handles) != 1 {
		t.Errorf("expected 1 handle, got %d", len(handles))
	}
}
