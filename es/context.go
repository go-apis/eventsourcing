package es

import (
	"context"
	"fmt"
	"net/http"
)

type Key int

const (
	NamespaceKey Key = iota
	UnitKey
	TxKey
)

const defaultNamespace = "default"

var ErrUnitNotFound = fmt.Errorf("unit not found")

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
func GetUnit(ctx context.Context) (Unit, error) {
	unit, ok := ctx.Value(UnitKey).(Unit)
	if ok {
		return unit, nil
	}
	return nil, ErrUnitNotFound
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

func CreateUnit(cli Client) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			unit, err := cli.Unit(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			ctx = SetUnit(ctx, unit)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
