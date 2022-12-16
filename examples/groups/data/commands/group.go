package commands

import (
	"github.com/contextcloud/eventstore/es"
	"github.com/google/uuid"
)

type CreateGroup struct {
	es.BaseCommand

	Name     string
	ParentId *uuid.UUID
}

type AddUser struct {
	es.BaseCommand

	UserId uuid.UUID
}
