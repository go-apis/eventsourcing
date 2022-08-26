package es

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/trace"
)

type Key int

const (
	NamespaceKey Key = iota
	UnitKey
	UserKey
)

const defaultNamespace = "default"

var ErrUnitNotFound = fmt.Errorf("unit not found")

func NamespaceFromContext(ctx context.Context) string {
	namespace, ok := ctx.Value(NamespaceKey).(string)
	if ok {
		return namespace
	}
	return defaultNamespace
}
func UserFromContext(ctx context.Context) User {
	user, ok := ctx.Value(UnitKey).(User)
	if ok {
		return user
	}
	return nil
}
func MetadataFromContext(ctx context.Context) map[string]interface{} {
	m := make(map[string]interface{})

	span := trace.SpanFromContext(ctx)
	if span != nil && span.SpanContext().HasSpanID() {
		m["span.span_id"] = span.SpanContext().SpanID().String()
	}
	if span != nil && span.SpanContext().HasTraceID() {
		m["span.trace_id"] = span.SpanContext().TraceID().String()
	}

	user := UserFromContext(ctx)
	if user != nil {
		m["user.id"] = user.Id().String()
	}
	return m
}

func GetUnit(ctx context.Context) (Unit, error) {
	unit, ok := ctx.Value(UnitKey).(Unit)
	if ok {
		return unit, nil
	}
	return nil, ErrUnitNotFound
}

func SetNamespace(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, NamespaceKey, namespace)
}
func SetUnit(ctx context.Context, unit Unit) context.Context {
	return context.WithValue(ctx, UnitKey, unit)
}
func SetUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, UserKey, user)
}
