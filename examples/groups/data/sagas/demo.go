package sagas

import (
	"context"
	"fmt"

	"github.com/go-apis/eventsourcing/examples/groups/data/events"

	"github.com/go-apis/eventsourcing/es"
)

type demoSaga struct {
	es.BaseSaga
}

func (s *demoSaga) HandleDemoCreated(ctx context.Context, evt *es.Event, data *events.DemoCreated) ([]es.Command, error) {
	fmt.Println("Demo created:", data.Name)
	return nil, nil
}

func NewDemoSaga() es.IsSaga {
	return &demoSaga{}
}
