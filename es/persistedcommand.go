package es

import (
	"time"

	"github.com/google/uuid"
)

type PersistedCommand struct {
	Id           uuid.UUID `json:"id" format:"uuid" required:"true"`
	Namespace    string    `json:"namespace" required:"true"`
	Command      Command   `json:"command" required:"true"`
	CommandType  string    `json:"command_type" required:"true"`
	ExecuteAfter time.Time `json:"execute_after" required:"true"`
	CreatedAt    time.Time `json:"created_at" required:"true"`
}
