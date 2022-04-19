package es

type EventStore interface {
	Connect() (Client, error)
}

type eventStore struct {
	opts options
}

func (s *eventStore) Connect() (Client, error) {
	return NewClient(), nil
}

func NewEventStore(opt ...Option) EventStore {
	opts := defaultOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	// build commandhandlers

	return &eventStore{
		opts: opts,
	}
}
