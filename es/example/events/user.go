package events

import "eventstore/es/types"

type UserCreated struct {
	Username string
	Password types.Encrypted
}
