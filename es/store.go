package es

import "context"

type Store interface {
	NewTx(ctx context.Context) (Tx, error)
	Load(ctx context.Context, id string, typeName string, out interface{}) error
	Save(ctx context.Context, id string, typeName string, out interface{}) ([]Event, error)
	GetEvents(ctx context.Context) ([]Event, error)
}

type store struct {
	inner Store
}

func (s *store) NewTx(ctx context.Context) (Tx, error) {
	return nil, nil
}

func (s *store) Load(ctx context.Context, id string, typeName string, out interface{}) error {
	return nil
}

func (s *store) Save(ctx context.Context, id string, typeName string, out interface{}) ([]Event, error) {
	return nil, nil
}

func (s *store) GetEvents(ctx context.Context) ([]Event, error) {
	return nil, nil
}

func NewStore(url string, serviceName string) (Store, error) {
	inner, err := NewDbStore(url, serviceName)
	if err != nil {
		return nil, err
	}

	return &store{
		inner: inner,
	}, nil
}
