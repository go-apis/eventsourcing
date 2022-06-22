package events

import (
	"github.com/contextcloud/eventstore/es/example/models"
	"github.com/contextcloud/eventstore/es/types"
)

type UserCreated struct {
	Username string
	Password types.Encrypted
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
