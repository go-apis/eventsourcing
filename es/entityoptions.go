package es

// EntityOptions represents the configuration options
// for the entity.
type EntityOptions struct {
	Name           string
	Factory        EntityFunc
	Revision       string
	MinVersionDiff int
	Project        bool
}

// EntityOption applies an option to the provided configuration.
type EntityOption func(*EntityOptions)

func EntityRevision(revision string) EntityOption {
	return func(o *EntityOptions) {
		o.Revision = revision
	}
}
func EntityRevisionMin(minVersionDiff int) EntityOption {
	return func(o *EntityOptions) {
		o.MinVersionDiff = minVersionDiff
	}
}
func EntityDisableRevision() EntityOption {
	return func(o *EntityOptions) {
		o.MinVersionDiff = -1
	}
}
func EntityDisableProject() EntityOption {
	return func(o *EntityOptions) {
		o.Project = false
	}
}
func EntityName(name string) EntityOption {
	return func(o *EntityOptions) {
		o.Name = name
	}
}
func EntityFactory(factory EntityFunc) EntityOption {
	return func(o *EntityOptions) {
		o.Factory = factory
	}
}

func NewEntityOptions(options []EntityOption) EntityOptions {
	// set defaults.
	o := EntityOptions{
		Revision:       "rev1",
		Project:        true,
		MinVersionDiff: 0,
	}

	// apply options.
	for _, opt := range options {
		opt(&o)
	}

	if o.Factory == nil {
		panic("You need to supply a factory method")
	}
	return o
}
