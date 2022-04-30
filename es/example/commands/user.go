package commands

import "eventstore/es"

type CreateUser struct {
	es.BaseCommand

	Username string
	Password string
}

type AddEmail struct {
	es.BaseCommand

	Email string
}
