package events

import (
	"github.com/go-apis/eventsourcing/es"
	"github.com/go-apis/eventsourcing/es/types"
	"github.com/go-apis/eventsourcing/examples/groups/models"
	"github.com/google/uuid"
)

type GroupCreated struct {
	Name     string
	ParentId *uuid.UUID
}

type UserAdded struct {
	Users types.SliceItem[models.User]
}

type GroupAdded struct {
	es.BaseEvent `es:"service=users"`

	Groups types.SliceItem[models.Group]
}
