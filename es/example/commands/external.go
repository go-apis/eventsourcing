package commands

import "eventstore/es"

type CreateExternalUser struct {
	es.BaseCommand

	Name     string
	UserId   string
	Username string
}
