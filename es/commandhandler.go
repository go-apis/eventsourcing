package es

import (
	"context"
)

type CommandHandler interface {
	Handle(ctx context.Context, cmd Command) error
}
