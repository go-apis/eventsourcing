package es

import "context"

type Store interface {
	Load(ctx context.Context, id string, typeName string, out interface{}) error
	Save(ctx context.Context, id string, typeName string, out interface{}) ([]Event, error)

	GetEvents(ctx context.Context) ([]Event, error)
}

func NewStore(url string, serviceName string) (Store, error) {
	// todo support different types of stores

	return NewDbStore(url, serviceName)
}
