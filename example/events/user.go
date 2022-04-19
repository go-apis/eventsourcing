package events

import "eventstore/example/es/types"

type UserCreated struct {
	Username string
	Password types.Encrypted
}
