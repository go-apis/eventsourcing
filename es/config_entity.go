package es

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jinzhu/copier"
)

// EntityConfig represents the configuration options
// for the entity.
type EntityConfig struct {
	Name             string
	Type             reflect.Type
	Factory          EntityFunc
	SnapshotRevision string
	SnapshotEnabled  bool
	SnapshotEvery    int
	Project          bool
	Handles          EventHandles
}

// EntityOption applies an option to the provided configuration.
type EntityOption func(*EntityConfig)

func NewEntityOptionsFromTag(t reflect.Type) ([]EntityOption, error) {
	field, ok := t.FieldByName("BaseAggregateSourced")
	if !ok {
		return nil, nil
	}
	tag := field.Tag.Get("es")
	if tag == "" {
		return nil, nil
	}

	var options []EntityOption

	// parse fields
	items := strings.Split(tag, ",")
	for _, item := range items {
		split := strings.Split(item, "=")
		part1 := split[0]

		if len(split) == 1 {
			if len(part1) > 0 {
				options = append(options, EntityName(part1))
			}
			continue
		}

		switch part1 {
		case "rev":
			options = append(options, EntitySnapshotRevision(split[1]))
			continue
		case "snapshot":
			i, err := strconv.Atoi(split[1])
			if err != nil {
				return nil, err
			}
			options = append(options, EntitySnapshotEvery(i))
			continue
		case "project":
			if split[1] == "false" {
				options = append(options, EntityDisableProject())
			}
			continue
		}
	}
	return options, nil
}

func NewEntityOptions(agg interface{}) []EntityOption {
	var t reflect.Type

	switch raw := agg.(type) {
	case reflect.Type:
		t = raw
		break
	default:
		t = reflect.TypeOf(agg)
	}

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	name := t.Name()
	factory := func() (Entity, error) {
		out := reflect.New(t).Interface().(Entity)
		if agg == nil {
			return out, nil
		}

		if err := copier.Copy(out, agg); err != nil {
			return nil, err
		}
		return out, nil
	}

	// read tag.
	tags, err := NewEntityOptionsFromTag(t)
	if err != nil {
		panic(err)
	}

	return append([]EntityOption{
		EntityType(t),
		EntityName(name),
		EntityFactory(factory),
	}, tags...)
}

func EntitySnapshotRevision(snapshotRevision string) EntityOption {
	return func(o *EntityConfig) {
		o.SnapshotRevision = snapshotRevision
	}
}
func EntitySnapshotEvery(versions int) EntityOption {
	return func(o *EntityConfig) {
		o.SnapshotEvery = versions
		o.SnapshotEnabled = true
	}
}
func EntityDisableSnapshot() EntityOption {
	return func(o *EntityConfig) {
		o.SnapshotEvery = -1
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

func NewEntityConfig(options []EntityOption) (*EntityConfig, error) {
	// set defaults.
	o := &EntityConfig{
		SnapshotRevision: "rev1",
		Project:          true,
		SnapshotEvery:    0,
		SnapshotEnabled:  false,
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
