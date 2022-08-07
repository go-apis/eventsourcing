package es

// UnitLoadOptions represents the configuration options for loading
type DataLoadOptions struct {
	Force bool
}

// DataLoadOption applies an option to the provided configuration.
type DataLoadOption func(*DataLoadOptions)

func DataLoadForce(force bool) DataLoadOption {
	return func(o *DataLoadOptions) {
		o.Force = force
	}
}
