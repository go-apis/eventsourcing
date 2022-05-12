package es

type options struct {
	commandHandlers []interface{}
	eventHandlers   []interface{}
}

type optionFunc struct {
	f func(*options)
}

func (of optionFunc) apply(o *options) {
	of.f(o)
}

func newOptionFunc(f func(*options)) *optionFunc {
	return &optionFunc{
		f: f,
	}
}

var defaultOptions = options{}

// Setup the options for the eventstore
type Option interface {
	apply(*options)
}

func WithCommandHandlers(handlers ...interface{}) Option {
	return newOptionFunc(func(o *options) {
		o.commandHandlers = append(o.commandHandlers, handlers...)
	})
}

func WithEventHandlers(handlers ...interface{}) Option {
	return newOptionFunc(func(o *options) {
		o.eventHandlers = append(o.eventHandlers, handlers...)
	})
}
