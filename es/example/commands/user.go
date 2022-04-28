package commands

type CreateUser struct {
	Username string
	Password string
}

func (c *CreateUser) GetAggregateId() string {
	return "user-1"
}
