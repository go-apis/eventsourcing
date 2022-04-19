package commands

type CreateUser struct {
	Username string
	Password string
}

func (c *CreateUser) AggregateId() string {
	return "user-1"
}
