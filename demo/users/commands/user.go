package commands

import (
	"github.com/contextcloud/eventstore/es"
	"github.com/google/uuid"
)

type CreateUser struct {
	es.BaseCommand

	Username string
	Password string
}

type AddEmail struct {
	es.BaseCommand

	Email string
}

type AddConnection struct {
	es.BaseCommand

	Name     string
	UserId   string
	Username string
}

type UpdateConnection struct {
	es.BaseCommand

	Username string
}

type AddGroup struct {
	es.BaseCommand

	GroupId uuid.UUID
	Name    string
}
