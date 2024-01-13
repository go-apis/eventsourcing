package es

import (
	"time"

	"github.com/google/uuid"
)

func Commands(cmds ...Command) []Command {
	return cmds
}

type Command interface {
	GetAggregateId() uuid.UUID
}

type ScheduledCommand interface {
	Command

	ExecuteAfter() time.Time
	GetCommand() Command
}

type ReplayCommand interface {
	GetAggregateName() string
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

type scheduledCommand struct {
	Command

	t time.Time
}

func (c *scheduledCommand) ExecuteAfter() time.Time {
	return c.t
}

func (c *scheduledCommand) GetCommand() Command {
	return c.Command
}

func NewScheduledCommand(cmd Command, t time.Time) ScheduledCommand {
	return &scheduledCommand{
		Command: cmd,
		t:       t,
	}
}

type replayCommand struct {
	BaseCommand
	BaseNamespaceCommand

	AggregateName string
}

func (c replayCommand) GetAggregateName() string {
	return c.AggregateName
}

func NewReplayCommand(namespace string, aggregateId uuid.UUID, aggregateName string) ReplayCommand {
	return &replayCommand{
		BaseCommand: BaseCommand{
			AggregateId: aggregateId,
		},
		BaseNamespaceCommand: BaseNamespaceCommand{
			Namespace: namespace,
		},
		AggregateName: aggregateName,
	}
}

func IsReplayCommand(cmd Command) bool {
	// handle the command
	switch cmd.(type) {
	case ReplayCommand:
		return true
	default:
		return false
	}
}
