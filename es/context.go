package es

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type Key int

const (
	NamespaceKey Key = iota
	UnitKey
	ActorKey
	SkipPublishKey
	TimeKey
	ClientKey
)

const defaultNamespace = "default"

var ErrUnitNotFound = fmt.Errorf("unit not found")

func GetNamespace(ctx context.Context) string {
	namespace, ok := ctx.Value(NamespaceKey).(string)
	if ok {
		return namespace
	}
	return defaultNamespace
}
func GetActor(ctx context.Context) *Actor {
	actor, ok := ctx.Value(ActorKey).(*Actor)
	if ok {
		return actor
	}
	return nil
}
func GetMetadata(ctx context.Context) map[string]interface{} {
	m := make(map[string]interface{})

	span := trace.SpanFromContext(ctx)
	if span != nil && span.SpanContext().HasSpanID() {
		m["span.span_id"] = span.SpanContext().SpanID().String()
	}
	if span != nil && span.SpanContext().HasTraceID() {
		m["span.trace_id"] = span.SpanContext().TraceID().String()
	}
	return m
}
func GetUnit(ctx context.Context) (Unit, error) {
	unit, ok := ctx.Value(UnitKey).(Unit)
	if ok {
		return unit, nil
	}
	return nil, ErrNotFound
}
func GetSkipPublish(ctx context.Context) bool {
	skip, ok := ctx.Value(SkipPublishKey).(bool)
	return ok && skip
}
func GetTime(ctx context.Context) time.Time {
	t, ok := ctx.Value(TimeKey).(time.Time)
	if ok {
		return t
	}
	return time.Now()
}

func SetNamespace(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, NamespaceKey, namespace)
}
func SetUnit(ctx context.Context, unit Unit) context.Context {
	return context.WithValue(ctx, UnitKey, unit)
}

// SetClient makes the owning client reachable from handler contexts so
// NewDetachedUnit can mint fresh units there.
func SetClient(ctx context.Context, cli Client) context.Context {
	return context.WithValue(ctx, ClientKey, cli)
}

func GetClient(ctx context.Context) (Client, error) {
	cli, ok := ctx.Value(ClientKey).(Client)
	if ok {
		return cli, nil
	}
	return nil, ErrNotFound
}

// NewDetachedUnit returns a fresh unit with its own transaction boundary,
// ignoring any unit already in ctx. Bus handlers use it to commit work in
// increments — one unit per chunk — instead of accumulating everything in
// the delivery-wide transaction: committed chunks survive a crash, update
// read models (and the outbox) as they land, and idempotent commands let a
// redelivery converge over them. Dispatch through it with
// es.SetUnit(ctx, unit) so nested handling stays inside the chunk.
//
// This is a Postgres pattern: providers pinned to one connection (sqlite
// :memory:) cannot begin a detached transaction while the delivery
// transaction holds the connection.
//
// Outside a bus delivery no client is registered — replay endpoints and
// tests invoke handlers on their own unit — so NewDetachedUnit falls back
// to the ctx unit there: chunked dispatch degrades to the caller's single
// transaction instead of failing.
func NewDetachedUnit(ctx context.Context) (Unit, error) {
	cli, err := GetClient(ctx)
	if err != nil {
		return GetUnit(ctx)
	}
	// Clear the unit key so the client mints a new one instead of
	// returning the delivery's unit.
	return cli.Unit(context.WithValue(ctx, UnitKey, nil))
}
func SetActor(ctx context.Context, actor *Actor) context.Context {
	return context.WithValue(ctx, ActorKey, actor)
}
func SetSkipPublish(ctx context.Context) context.Context {
	return context.WithValue(ctx, SkipPublishKey, true)
}
func SetTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, TimeKey, t)
}
