package sagas

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/examples/groups/data/events"

	"github.com/contextcloud/eventstore/es"
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
