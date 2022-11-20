package es

import "github.com/google/uuid"

func Commands(cmds ...Command) []Command {
	return cmds
}

type Command interface {
	GetAggregateId() uuid.UUID
}

type CommandPerms interface {
	GetPerms() string
}

type NamespaceCommand interface {
	GetNamespace() string
}

type BaseNamespaceCommand struct {
	Namespace string `json:"namespace" required:"true"`
}

// GetNamespace return the namespace for the command
func (c BaseNamespaceCommand) GetNamespace() string {
	return c.Namespace
}

// BaseCommand to make it easier to get the ID
type BaseCommand struct {
	AggregateId uuid.UUID `json:"aggregate_id" format:"uuid" required:"true"`
}

// GetAggregateId return the aggregate id
func (c BaseCommand) GetAggregateId() uuid.UUID {
	return c.AggregateId
}

// ReplayCommand a command that load and reply events ontop of an aggregate.
type ReplayCommand struct {
	BaseCommand
	BaseNamespaceCommand

	AggregateName string
}

func IsReplayCommand(cmd Command) bool {
	// handle the command
	switch cmd.(type) {
	case ReplayCommand:
		return true
	case *ReplayCommand:
		return true
	default:
		return false
	}
}
