package es

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

var ErrHandlerNotFound = fmt.Errorf("handler not found")
var ErrNotCommandHandler = fmt.Errorf("not a command handler")

type Unit interface {
	GetData() Data
	NewTx(ctx context.Context) (Tx, error)

	CreateCommand(name string) (Command, error)
	Dispatch(ctx context.Context, cmds ...Command) error
	DispatchAsync(ctx context.Context, cmds ...Command) error

	Load(ctx context.Context, name string, id uuid.UUID, dataOptions ...DataLoadOption) (Entity, error)
	Save(ctx context.Context, name string, entity Entity) error
}

type unit struct {
	sync.RWMutex

	cli  Client
	data Data
	tx   Tx

	events []*Event
}

func (u *unit) GetData() Data {
	return u.data
}

func (u *unit) NewTx(ctx context.Context) (Tx, error) {
	u.Lock()
	defer u.Unlock()

	if u.tx != nil {
		return u, nil
	}

	tx, err := u.data.Begin(ctx)
	if err != nil {
		return nil, err
	}

	u.tx = tx
	return u, nil
}

func (u *unit) CreateCommand(name string) (Command, error) {
	// create the command.
	cfg, err := u.cli.GetCommandConfig(name)
	if err != nil {
		return nil, err
	}
	return cfg.Factory()
}

func (u *unit) Dispatch(ctx context.Context, cmds ...Command) error {
	ctx = SetUnit(ctx, u)
	return u.cli.HandleCommands(ctx, cmds...)
}
func (u *unit) DispatchAsync(ctx context.Context, cmds ...Command) error {
	ctx = SetUnit(ctx, u)
	return u.cli.HandleCommandsAsync(ctx, cmds...)
}

func (u *unit) Commit(ctx context.Context) (int, error) {
	u.Lock()
	defer u.Unlock()

	if u.tx == nil {
		return 0, nil
	}
	out, err := u.tx.Commit(ctx)
	if err != nil {
		return out, err
	}
	u.tx = nil

	// send over the
	if err := u.cli.PublishEvents(ctx, u.events...); err != nil {
		// TODO log this!!!
		return 0, err
	}
	u.events = nil
	return out, nil
}

func (u *unit) Rollback(ctx context.Context) error {
	u.Lock()
	defer u.Unlock()

	if u.tx == nil {
		return nil
	}
	err := u.tx.Rollback(ctx)
	if err != nil {
		return err
	}
	u.tx = nil
	return nil
}

func (u *unit) Load(ctx context.Context, name string, id uuid.UUID, dataOptions ...DataLoadOption) (Entity, error) {
	entityConfig, err := u.cli.GetEntityConfig(name)
	if err != nil {
		return nil, err
	}
	dataStore := NewDataStore(u.data, entityConfig)
	return dataStore.Load(ctx, id, dataOptions...)
}
func (u *unit) Save(ctx context.Context, name string, entity Entity) error {
	entityConfig, err := u.cli.GetEntityConfig(name)
	if err != nil {
		return err
	}

	dataStore := NewDataStore(u.data, entityConfig)
	events, err := dataStore.Save(ctx, entity)
	if err != nil {
		return err
	}
	u.events = append(u.events, events...)
	return u.cli.HandleEvents(ctx, events...)
}

func newUnit(cli Client, data Data) (Unit, error) {
	return &unit{
		cli:  cli,
		data: data,
	}, nil
}
