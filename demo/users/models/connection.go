package models

type Connection struct {
	Name     string
	UserId   string
	Username string
}

type ConnectionUpdate struct {
	Username string
}
