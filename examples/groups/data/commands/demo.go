package commands

import "github.com/go-apis/eventsourcing/es"

type NewDemo struct {
	es.BaseCommand

	Name string `json:"name"`
}
