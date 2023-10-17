package es

import "context"

type demoEventData struct {
	Something string
}

type demoSaga struct {
	BaseSaga
}

func (s *demoSaga) HandleEventData(ctx context.Context, evt *Event, data *demoEventData) ([]Command, error) {
	return nil, nil
}

type demoEntity struct {
	BaseAggregate

	Something string
}

type demoProjector struct {
	BaseProjector
}

func (s *demoProjector) HandleEventData(ctx context.Context, agg *demoEntity, data *demoEventData) error {
	agg.Something = data.Something
	return nil
}
