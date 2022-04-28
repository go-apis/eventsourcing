package es

import "context"

type Store interface {
	Load(ctx context.Context, id string, typeName string, out interface{}) error
}

func NewStore(url string) (Store, error) {
	// todo support different types of stores

	return NewDbStore(url)
}
