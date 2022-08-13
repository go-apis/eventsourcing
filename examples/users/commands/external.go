package commands

import "github.com/contextcloud/eventstore/es"

type CreateExternalUser struct {
	es.BaseCommand

	Name     string
	UserId   string
	Username string
}
