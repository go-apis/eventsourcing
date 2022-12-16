package aggregates

import "github.com/contextcloud/eventstore/es"

type MyName struct {
	es.BaseAggregateHolder

	Name string `json:"name"`
}
