package es

type QueryOptions struct {
	UseNamespace bool
	Namespace    string
}

type QueryOption func(*QueryOptions)

func WithNoNamespace() QueryOption {
	return func(opts *QueryOptions) {
		opts.UseNamespace = false
	}
}

func WithNamespace(namespace string) QueryOption {
	return func(opts *QueryOptions) {
		opts.Namespace = namespace
	}
}

func DefaultQueryOptions() *QueryOptions {
	return &QueryOptions{
		UseNamespace: true,
		Namespace:    "",
	}
}
