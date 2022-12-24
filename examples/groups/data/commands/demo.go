package commands

import "github.com/contextcloud/eventstore/es"

type NewDemo struct {
	es.BaseCommand

	Name string `json:"name"`
}
