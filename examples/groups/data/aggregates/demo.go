package aggregates

import "github.com/contextcloud/eventstore/es"

type Demo struct {
	es.BaseAggregateHolder

	Name string `json:"name"`
}
