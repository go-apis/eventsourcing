package aggregates

import "github.com/go-apis/eventsourcing/es"

type Demo struct {
	es.BaseAggregateHolder

	Name string `json:"name"`
}
