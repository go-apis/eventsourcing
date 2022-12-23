package es

// AggregateHolder for event stored aggregates
type AggregateHolder interface {
	Entity

	// Events returns events that need to be published
	GetEvents() []*Event
}

// BaseAggregateHolder to make our commands smaller
type BaseAggregateHolder struct {
	BaseAggregate

	events []interface{}
}

// PublishEvent registers an event to be published after the aggregate
// has been successfully saved.
func (a *BaseAggregateHolder) Publish(e interface{}) {
	a.events = append(a.events, e)
}

// Events returns all uncommitted events that are not yet saved.
func (a *BaseAggregateHolder) GetEvents() []interface{} {
	return a.events
}
