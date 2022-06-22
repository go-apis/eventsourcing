package es

import (
	"context"
)

type Key int

const (
	NamespaceKey Key = iota
	UnitKey
	TxKey
)

const defaultNamespace = "default"

func SetNamespace(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, NamespaceKey, namespace)
}

func NamespaceFromContext(ctx context.Context) string {
	namespace, ok := ctx.Value(NamespaceKey).(string)
	if ok {
		return namespace
	}
	return defaultNamespace
}

func SetTx(ctx context.Context, tx Tx) context.Context {
	return context.WithValue(ctx, TxKey, tx)
}
func TxFromContext(ctx context.Context) Tx {
	tx, ok := ctx.Value(TxKey).(Tx)
	if ok {
		return tx
	}
	return nil
}

func SetUnit(ctx context.Context, unit Unit) context.Context {
	return context.WithValue(ctx, UnitKey, unit)
}
func UnitFromContext(ctx context.Context) Unit {
	unit, ok := ctx.Value(UnitKey).(Unit)
	if ok {
		return unit
	}
	return nil
}

func MetadataFromContext(ctx context.Context) map[string]interface{} {
	m := make(map[string]interface{})

	// if md, ok := metadata.FromIncomingContext(ctx); ok {
	// 	for k, v := range md {
	// 		if len(v) == 1 {
	// 			m[k] = v[0]
	// 			continue
	// 		}
	// 		m[k] = v
	// 	}
	// }

	// what about tracing?

	return m
}
