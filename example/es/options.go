package es

type options struct {
	url      string
	entities []interface{}
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

var defaultOptions = options{
	url: "http://localhost:6632",
}

// Setup the options for the eventstore
type Option interface {
	apply(*options)
}

func WithEntities(entities ...interface{}) Option {
	return newOptionFunc(func(o *options) {
		o.entities = append(o.entities, entities...)
	})
}
