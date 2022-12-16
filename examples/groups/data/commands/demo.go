package commands

import "github.com/contextcloud/eventstore/es"

type DemoCommand struct {
	es.BaseCommand

	Name string
}
