package es

type QueryOptions struct {
	UseNamespace bool
}

type QueryOption func(*QueryOptions)

func WithNoNamespace() QueryOption {
	return func(opts *QueryOptions) {
		opts.UseNamespace = false
	}
}

func DefaultQueryOptions() *QueryOptions {
	return &QueryOptions{
		UseNamespace: true,
	}
}
