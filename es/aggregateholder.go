package es

import "github.com/google/uuid"

// AggregateHolder for event stored aggregates
type AggregateHolder interface {
	Entity

	// EventsToPublish returns events that need to be published
	EventsToPublish() []*Event
}

// BaseAggregateHolder to make our commands smaller
type BaseAggregateHolder struct {
	Id        uuid.UUID `json:"id"`
	Namespace string    `json:"namespace"`

	events []Event
}

// GetId of the aggregate
func (a *BaseAggregateHolder) GetId() uuid.UUID {
	return a.Id
}

// SetId of the aggregate
func (a *BaseAggregateHolder) SetId(id uuid.UUID, namespace string) {
	a.Id = id
	a.Namespace = namespace
}

// PublishEvent registers an event to be published after the aggregate
// has been successfully saved.
func (a *BaseAggregateHolder) PublishEvent(e Event) {
	a.events = append(a.events, e)
}

// Events returns all uncommitted events that are not yet saved.
func (a *BaseAggregateHolder) EventsToPublish() []Event {
	return a.events
}
