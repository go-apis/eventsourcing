package es

type options struct {
	url      string
	handlers []interface{}
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

func WithHandlers(handlers ...interface{}) Option {
	return newOptionFunc(func(o *options) {
		o.handlers = append(o.handlers, handlers...)
	})
}

func WithDb() Option {
	return newOptionFunc(func(o *options) {
		o.url = "http://localhost:6632"
	})
}
