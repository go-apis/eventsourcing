package es

import (
	"fmt"
	"reflect"

	"github.com/jinzhu/copier"
)

// EntityConfig represents the configuration options
// for the entity.
type EntityConfig struct {
	Name           string
	Type           reflect.Type
	Factory        EntityFunc
	Mapper         EventDataMapper
	Revision       string
	MinVersionDiff int
	Project        bool
}

// EntityOption applies an option to the provided configuration.
type EntityOption func(*EntityConfig)

func NewEntityOptions(agg interface{}) []EntityOption {
	t := reflect.TypeOf(agg)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	name := t.Name()
	factory := func() (Entity, error) {
		out := reflect.New(t).Interface().(Entity)
		if err := copier.Copy(out, agg); err != nil {
			return nil, err
		}
		return out, nil
	}
	return []EntityOption{
		EntityType(t),
		EntityName(name),
		EntityFactory(factory),
	}
}

func EntityRevision(revision string) EntityOption {
	return func(o *EntityConfig) {
		o.Revision = revision
	}
}
func EntityRevisionMin(minVersionDiff int) EntityOption {
	return func(o *EntityConfig) {
		o.MinVersionDiff = minVersionDiff
	}
}
func EntityDisableRevision() EntityOption {
	return func(o *EntityConfig) {
		o.MinVersionDiff = -1
	}
}
func EntityDisableProject() EntityOption {
	return func(o *EntityConfig) {
		o.Project = false
	}
}
func EntityName(name string) EntityOption {
	return func(o *EntityConfig) {
		o.Name = name
	}
}
func EntityFactory(factory EntityFunc) EntityOption {
	return func(o *EntityConfig) {
		o.Factory = factory
	}
}
func EntityType(t reflect.Type) EntityOption {
	return func(o *EntityConfig) {
		o.Type = t
	}
}
func EntityEventTypes(objs ...interface{}) EntityOption {
	mapper := NewEventDataFunc(objs...)

	return func(o *EntityConfig) {
		for k, v := range mapper {
			o.Mapper[k] = v
		}
	}
}

func NewEntityConfig(options []EntityOption) (*EntityConfig, error) {
	// set defaults.
	o := &EntityConfig{
		Revision:       "rev1",
		Project:        true,
		MinVersionDiff: 0,
		Mapper:         make(EventDataMapper),
	}

	// apply options.
	for _, opt := range options {
		opt(o)
	}

	if o.Factory == nil {
		return nil, fmt.Errorf("factory is required")
	}
	if o.Type == nil {
		return nil, fmt.Errorf("type is required")
	}
	if o.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	return o, nil
}
