package events

import (
	"github.com/go-apis/eventsourcing/es"
	"github.com/go-apis/eventsourcing/es/types"
	"github.com/go-apis/eventsourcing/examples/users/models"
)

type UserCreated struct {
	Username string
	Password string
}
type UserDeleted struct {
	Deleted bool
}

type EmailAdded struct {
	Email string
}

type ConnectionAdded struct {
	Connections types.SliceItem[models.Connection]
}

type ConnectionUpdated struct {
	Connections types.SliceItem[models.ConnectionUpdate]
}

type GroupAdded struct {
	es.BaseEvent `es:"publish"`

	Groups types.SliceItem[models.Group]
}
