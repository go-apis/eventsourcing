package es

import "github.com/google/uuid"

type User interface {
	Id() uuid.UUID
}
