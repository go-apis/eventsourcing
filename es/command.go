package es

import "github.com/google/uuid"

type Command interface {
	GetAggregateId() uuid.UUID
}

// BaseCommand to make it easier to get the ID
type BaseCommand struct {
	AggregateId uuid.UUID `json:"aggregate_id"`
}

// GetAggregateId return the aggregate id
func (c BaseCommand) GetAggregateId() uuid.UUID {
	return c.AggregateId
}

// ReplayCommand a command that load and reply events ontop of an aggregate.
type ReplayCommand struct {
	AggregateId uuid.UUID `json:"aggregate_id"`
}

// GetAggregateId return the aggregate id
func (c ReplayCommand) GetAggregateId() uuid.UUID {
	return c.AggregateId
}

func IsReplayCommand(cmd Command) bool {
	// handle the command
	switch cmd.(type) {
	case *ReplayCommand:
		return true
	default:
		return false
	}
}
