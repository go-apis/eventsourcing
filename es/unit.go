package es

import (
	"context"
	"fmt"
	"sync"

	"github.com/contextcloud/eventstore/es/filters"
	"github.com/google/uuid"
)

type Unit interface {
	Get(ctx context.Context, aggregateName string, namespace string, id uuid.UUID, out interface{}) error
	Find(ctx context.Context, aggregateName string, namespace string, filter filters.Filter, out interface{}) error
	Count(ctx context.Context, aggregateName string, namespace string, filter filters.Filter) (int, error)

	Load(ctx context.Context, name string, id uuid.UUID, opts ...DataLoadOption) (Entity, error)
	Save(ctx context.Context, name string, aggregate Entity) error
	Delete(ctx context.Context, name string, aggregate Entity) error
	Truncate(ctx context.Context, name string) error

	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error

	Handle(ctx context.Context, events ...*Event) error
	Dispatch(ctx context.Context, cmds ...Command) error
	Replay(ctx context.Context, cmds ...*ReplayCommand) error
}

type unit struct {
	sync.RWMutex

	data      Data
	dataStore DataStore
	publisher Streamer

	tx     Tx
	events []*Event
}

func (u *unit) Get(ctx context.Context, aggregateName string, namespace string, id uuid.UUID, out interface{}) error {
	return u.data.Get(ctx, aggregateName, namespace, id, out)
}

func (u *unit) Find(ctx context.Context, aggregateName string, namespace string, filter filters.Filter, out interface{}) error {
	return u.data.Find(ctx, aggregateName, namespace, filter, out)
}

func (u *unit) Count(ctx context.Context, aggregateName string, namespace string, filter filters.Filter) (int, error) {
	return u.data.Count(ctx, aggregateName, namespace, filter)
}

func (u *unit) Load(ctx context.Context, name string, id uuid.UUID, opts ...DataLoadOption) (Entity, error) {
	return u.dataStore.Load(ctx, name, id, opts...)
}

func (u *unit) Save(ctx context.Context, name string, aggregate Entity) error {
	evts, err := u.dataStore.Save(ctx, name, aggregate)
	if err != nil {
		return err
	}

	// do something with events.
	u.events = append(u.events, evts...)

	for _, evt := range evts {
		handlers := GlobalRegistry.GetEventHandlers(evt.Data)
		for _, h := range handlers {
			if err := h.Handle(ctx, evt); err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *unit) Delete(ctx context.Context, name string, aggregate Entity) error {
	return u.dataStore.Delete(ctx, name, aggregate)
}

func (u *unit) Truncate(ctx context.Context, name string) error {
	return u.dataStore.Truncate(ctx, name)
}

func (u *unit) Commit(ctx context.Context) error {
	u.Lock()
	defer u.Unlock()

	if u.tx == nil {
		return nil
	}
	if err := u.tx.Commit(ctx); err != nil {
		return err
	}
	u.tx = nil

	if err := u.publisher.Publish(ctx, u.events...); err != nil {
		return err
	}

	u.events = nil
	return nil
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

func (u *unit) work(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	defer func() {
		if perr := recover(); perr != nil {
			err = fmt.Errorf("panic: %v", perr)
		}
		if err != nil {
			if rerr := u.Rollback(ctx); rerr != nil {
				err = fmt.Errorf("rolling back transaction fail: %s\n %w ", rerr.Error(), err)
			}
		}
	}()

	u.Lock()
	tx, err := u.data.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction fail: %w", err)
	}
	u.tx = tx
	u.Unlock()

	if err = fn(ctx); err != nil {
		return
	}

	if rerr := u.Commit(ctx); rerr != nil {
		return fmt.Errorf("committing transaction fail: %w", rerr)
	}

	return nil
}

func (u *unit) Handle(ctx context.Context, events ...*Event) error {
	ctx = SetUnit(ctx, u)

	return u.work(ctx, func(ctx context.Context) error {
		for _, evt := range events {
			// set the namespace
			nctx := SetNamespace(ctx, evt.Namespace)

			hs := GlobalRegistry.GetEventHandlers(evt.Data)
			if len(hs) == 0 {
				continue
			}
			for _, h := range hs {
				if err := h.Handle(nctx, evt); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
func (u *unit) Dispatch(ctx context.Context, cmds ...Command) error {
	ctx = SetUnit(ctx, u)

	return u.work(ctx, func(ctx context.Context) error {
		for _, cmd := range cmds {
			h, err := GlobalRegistry.GetCommandHandler(cmd)
			if err != nil {
				return err
			}
			if err := h.Handle(ctx, cmd); err != nil {
				return err
			}
		}
		return nil
	})
}
func (u *unit) Replay(ctx context.Context, cmds ...*ReplayCommand) error {
	ctx = SetUnit(ctx, u)

	return u.work(ctx, func(ctx context.Context) error {
		for _, cmd := range cmds {
			h, err := GlobalRegistry.GetReplayHandler(cmd)
			if err != nil {
				return err
			}
			if err := h.Handle(ctx, cmd); err != nil {
				return err
			}
		}
		return nil
	})
}

func newUnit(ctx context.Context, client *client) (Unit, error) {
	data, err := client.conn.NewData(ctx)
	if err != nil {
		return nil, err
	}

	ds := NewDataStore(client.providerConfig.Service, data)

	return &unit{
		data:      data,
		dataStore: ds,
		publisher: client.streamer,
	}, nil
}
