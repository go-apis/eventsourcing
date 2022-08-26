package events

import (
	"github.com/contextcloud/eventstore/es/types"
	"github.com/contextcloud/eventstore/examples/groups/models"
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
	Groups types.SliceItem[models.Group]
}
