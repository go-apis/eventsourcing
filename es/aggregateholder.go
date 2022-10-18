package es

// AggregateHolder for event stored aggregates
type AggregateHolder interface {
	Entity

	// EventsToPublish returns events that need to be published
	EventsToPublish() []*Event
}

// BaseAggregateHolder to make our commands smaller
type BaseAggregateHolder struct {
	BaseAggregate

	events []Event
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
