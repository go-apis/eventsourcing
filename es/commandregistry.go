package es

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-apis/eventsourcing/es/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var (
	// ErrHandlerAlreadySet is when a handler is already registered for a command.
	ErrHandlerAlreadySet = errors.New("handler is already set")
	// ErrHandlerNotFound is when no handler can be found.
	ErrHandlerNotFound = errors.New("no handlers for command")
)

type CommandRegistry interface {
	CommandHandler

	AddCommandConfig(cmdConfig *CommandConfig) error
	SetReplayHandler(h CommandHandler, entityConfig *EntityConfig) error
	SetCommandHandler(h CommandHandler, commandConfig *CommandConfig) error

	GetCommandConfig(name string) (*CommandConfig, error)
}

type commandRegistry struct {
	commands        []*CommandConfig
	hash            map[string]*CommandConfig
	commandHandlers map[string]CommandHandler
	replayHandlers  map[string]CommandHandler
}

func (r *commandRegistry) AddCommandConfig(cmdConfig *CommandConfig) error {
	if cmdConfig == nil {
		return errors.New("command config is nil")
	}

	lowered := strings.ToLower(cmdConfig.Name)
	if _, ok := r.hash[lowered]; ok {
		return fmt.Errorf("command %s already exists", lowered)
	}

	r.commands = append(r.commands, cmdConfig)
	r.hash[lowered] = cmdConfig
	return nil
}

func (r *commandRegistry) GetCommandConfig(name string) (*CommandConfig, error) {
	lowered := strings.ToLower(name)
	if cmdConfig, ok := r.hash[lowered]; ok {
		return cmdConfig, nil
	}

	return nil, errors.New("command not found")
}

func (r *commandRegistry) SetReplayHandler(h CommandHandler, entityConfig *EntityConfig) error {
	lowered := strings.ToLower(entityConfig.Name)
	if _, ok := r.replayHandlers[lowered]; ok {
		return ErrHandlerAlreadySet
	}
	r.replayHandlers[lowered] = h
	return nil
}

func (r *commandRegistry) SetCommandHandler(h CommandHandler, commandConfig *CommandConfig) error {
	lowered := strings.ToLower(commandConfig.Name)
	if _, ok := r.commandHandlers[lowered]; ok {
		return ErrHandlerAlreadySet
	}
	r.commandHandlers[lowered] = h
	return r.AddCommandConfig(commandConfig)
}

func (r *commandRegistry) HandleCommand(ctx context.Context, cmd Command) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	pctx, pspan := otel.Tracer("unit").Start(ctx, "dispatch")
	defer pspan.End()

	pspan.SetAttributes(
		attribute.String("id", cmd.GetAggregateId().String()),
	)

	// check if we have a namespace command
	if nsCmd, ok := cmd.(NamespaceCommand); ok {
		ns := nsCmd.GetNamespace()
		if ns != "" {
			pctx = SetNamespace(pctx, ns)
		}
	}

	replay, ok := cmd.(ReplayCommand)
	if ok {
		aggregateName := strings.ToLower(replay.GetAggregateName())
		if handler, ok := r.replayHandlers[aggregateName]; ok {
			return handler.HandleCommand(pctx, cmd)
		}
		return ErrHandlerNotFound
	}

	commandName := strings.ToLower(utils.GetTypeName(cmd))
	if handler, ok := r.commandHandlers[commandName]; ok {
		return handler.HandleCommand(pctx, cmd)
	}

	return ErrHandlerNotFound
}

func NewCommandRegistry() CommandRegistry {
	return &commandRegistry{
		commands:        []*CommandConfig{},
		hash:            map[string]*CommandConfig{},
		commandHandlers: map[string]CommandHandler{},
		replayHandlers:  map[string]CommandHandler{},
	}
}
