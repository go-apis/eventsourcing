package aggregates

import (
	"github.com/contextcloud/eventstore/es"
)

type User struct {
	es.BaseAggregate

	Type     string
	Username string
	Email    string
}
