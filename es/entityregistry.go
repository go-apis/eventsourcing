package es

import (
	"fmt"
	"strings"
)

type EntityRegistry interface {
	AddEntity(entityConfig *EntityConfig) error
	GetEntities() []*EntityConfig
	GetEntityConfig(name string) (*EntityConfig, error)
}

type entityRegistry struct {
	entities []*EntityConfig
	byname   map[string]*EntityConfig
}

func (r *entityRegistry) AddEntity(entityConfig *EntityConfig) error {
	if entityConfig == nil {
		return fmt.Errorf("entity config is nil")
	}

	name := entityConfig.Name
	name = strings.ToLower(name)

	if _, ok := r.byname[name]; ok {
		return fmt.Errorf("entity config already exists: %s", entityConfig.Name)
	}

	r.entities = append(r.entities, entityConfig)
	r.byname[name] = entityConfig
	return nil
}

func (r *entityRegistry) GetEntities() []*EntityConfig {
	return r.entities
}

func (r *entityRegistry) GetEntityConfig(name string) (*EntityConfig, error) {
	lower := strings.ToLower(name)
	if entityConfig, ok := r.byname[lower]; ok {
		return entityConfig, nil
	}
	return nil, fmt.Errorf("entity config not found: %s", name)
}

func NewEntityRegistry() EntityRegistry {
	return &entityRegistry{
		byname: make(map[string]*EntityConfig),
	}
}
