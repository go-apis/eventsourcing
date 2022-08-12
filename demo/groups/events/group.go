package events

import (
	"github.com/google/uuid"
)

type GroupCreated struct {
	Name     string
	ParentId *uuid.UUID
}

type UserAdded struct {
	Name   string
	UserId uuid.UUID
}
