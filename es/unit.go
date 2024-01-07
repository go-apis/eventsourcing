package es

import (
	"context"
	"fmt"
	"sync"

	"github.com/contextcloud/eventstore/es/filters"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type Unit interface {
	Get(ctx context.Context, aggregateName string, namespace string, id uuid.UUID, out interface{}) error
	Find(ctx context.Context, aggregateName string, namespace string, filter filters.Filter, out interface{}) error
	Count(ctx context.Context, aggregateName string, namespace string, filter filters.Filter) (int, error)

	Load(ctx context.Context, name string, id uuid.UUID, opts ...DataLoadOption) (Entity, error)
	Save(ctx context.Context, name string, aggregate Entity) error
	Delete(ctx context.Context, name string, aggregate Entity) error
	Truncate(ctx context.Context, name string) error

	FindEvents(ctx context.Context, filter filters.Filter) ([]*Event, error)

	Handle(ctx context.Context, group string, events ...*Event) error
	Dispatch(ctx context.Context, cmds ...Command) error
	Replay(ctx context.Context, cmds ...*ReplayCommand) error
}

type unit struct {
	sync.RWMutex

	registry  Registry
	data      Data
	dataStore DataStore
	publisher EventPublisher

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

	// TODO maybe move to the outbox pattern here.
	for _, evt := range evts {
		handlers := u.registry.GetEventHandlers(InternalGroup, evt.Data)
		if err := handlers.Handle(ctx, evt); err != nil {
			return err
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

func (u *unit) FindEvents(ctx context.Context, filter filters.Filter) ([]*Event, error) {
	return u.data.FindEvents(ctx, filter)
}

func (u *unit) work(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	tx, err := u.data.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction fail: %w", err)
	}

	skipPublish := GetSkipPublish(ctx)

	defer func() {
		if perr := recover(); perr != nil {
			err = fmt.Errorf("panic: %v", perr)
		}
		if err != nil {
			if rerr := tx.Rollback(ctx); rerr != nil {
				err = fmt.Errorf("rolling back transaction fail: %s\n %w ", rerr.Error(), err)
			}
		}
	}()

	sctx := SetSkipPublish(ctx)
	if err = fn(sctx); err != nil {
		return
	}

	if rerr := tx.Commit(ctx); rerr != nil {
		return fmt.Errorf("committing transaction fail: %w", rerr)
	}

	// publish events?
	if !skipPublish {
		for _, evt := range u.events {
			// can we publish it?
			cfg, err := u.registry.GetEventConfig(evt.Service, evt.Type)
			if err != nil || cfg.Publish == false {
				continue
			}

			if err := u.publisher.Publish(ctx, evt); err != nil {
				return err
			}
		}

		u.events = nil
	}

	return nil
}

func (u *unit) handle(ctx context.Context, group string, evt *Event) error {
	pctx, pspan := otel.Tracer("unit").Start(ctx, "handle")
	defer pspan.End()

	// set the namespace
	pctx = SetNamespace(pctx, evt.Namespace)

	hs := u.registry.GetEventHandlers(group, evt.Data)
	if len(hs) == 0 {
		return nil
	}

	pspan.SetAttributes(
		attribute.String("id", evt.AggregateId.String()),
	)

	for _, h := range hs {
		if err := h.Handle(pctx, evt); err != nil {
			return err
		}
	}
	return nil
}
func (u *unit) Handle(ctx context.Context, group string, events ...*Event) error {
	if len(events) == 0 {
		return nil
	}

	ctx = SetUnit(ctx, u)

	return u.work(ctx, func(ctx context.Context) error {
		for _, evt := range events {
			if err := u.handle(ctx, group, evt); err != nil {
				return err
			}
		}
		return nil
	})
}
func (u *unit) dispatch(ctx context.Context, cmd Command) error {
	pctx, pspan := otel.Tracer("unit").Start(ctx, "dispatch")
	defer pspan.End()

	h, err := u.registry.GetCommandHandler(cmd)
	if err != nil {
		return err
	}

	// check if we have a namespace command
	if nsCmd, ok := cmd.(NamespaceCommand); ok {
		ns := nsCmd.GetNamespace()
		if ns != "" {
			pctx = SetNamespace(pctx, ns)
		}
	}

	pspan.SetAttributes(
		attribute.String("id", cmd.GetAggregateId().String()),
	)

	if err := h.Handle(pctx, cmd); err != nil {
		return err
	}
	return nil
}
func (u *unit) Dispatch(ctx context.Context, cmds ...Command) error {
	if len(cmds) == 0 {
		return nil
	}

	ctx = SetUnit(ctx, u)

	return u.work(ctx, func(ctx context.Context) error {
		for _, cmd := range cmds {
			if err := u.dispatch(ctx, cmd); err != nil {
				return err
			}
		}
		return nil
	})
}
func (u *unit) replay(ctx context.Context, cmd *ReplayCommand) error {
	pctx, pspan := otel.Tracer("unit").Start(ctx, "dispatch")
	defer pspan.End()

	h, err := u.registry.GetReplayHandler(cmd)
	if err != nil {
		return err
	}

	ns := cmd.GetNamespace()
	if ns != "" {
		pctx = SetNamespace(pctx, ns)
	}

	pspan.SetAttributes(
		attribute.String("id", cmd.GetAggregateId().String()),
	)

	if err := h.Handle(pctx, cmd); err != nil {
		return err
	}
	return nil
}
func (u *unit) Replay(ctx context.Context, cmds ...*ReplayCommand) error {
	if len(cmds) == 0 {
		return nil
	}

	ctx = SetUnit(ctx, u)

	return u.work(ctx, func(ctx context.Context) error {
		for _, cmd := range cmds {
			if err := u.replay(ctx, cmd); err != nil {
				return err
			}
		}
		return nil
	})
}

func newUnit(ctx context.Context, service string, registry Registry, conn Conn, publisher EventPublisher) (Unit, error) {
	data, err := conn.NewData(ctx)
	if err != nil {
		return nil, err
	}

	ds := NewDataStore(service, data, registry)

	return &unit{
		data:      data,
		dataStore: ds,
		publisher: publisher,
	}, nil
}
