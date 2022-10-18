package es

import (
	"context"
)

type CommandHandler interface {
	Handle(ctx context.Context, cmd Command) error
}

// CommandHandlerFunc is a function that can be used as a command handler.
type CommandHandlerFunc func(context.Context, Command) error

// Handle is a method of the CommandHandler.
func (h CommandHandlerFunc) Handle(ctx context.Context, cmd Command) error {
	return h(ctx, cmd)
}
