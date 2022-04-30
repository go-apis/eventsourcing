package es

import (
	"context"
)

type key int

const namespaceKey key = 0

const defaultNamespace = "default"

func SetNamespace(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, namespaceKey, namespace)
}

func NamespaceFromContext(ctx context.Context) string {
	namespace, ok := ctx.Value(namespaceKey).(string)
	if ok {
		return namespace
	}
	return defaultNamespace
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
