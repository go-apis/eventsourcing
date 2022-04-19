package es

type Command interface {
	GetAggregateId() string
}

// BaseCommand to make it easier to get the ID
type BaseCommand struct {
	AggregateId string `json:"aggregate_id"`
}

// GetAggregateId return the aggregate id
func (c BaseCommand) GetAggregateId() string {
	return c.AggregateId
}
