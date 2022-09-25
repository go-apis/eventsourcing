package identity

import "fmt"

type SessionRequest struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type Identity struct {
	UserId     string                 `json:"user_id" validate:"required"`
	Username   string                 `json:"username" validate:"required"`
	Connection string                 `json:"provider" validate:"required"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// The user object
type User interface {
	GetRoles() []string
	GetIdentities() []Identity
	GetAudience() string
}

type ErrorMessage struct {
	// Error code
	Code int `json:"code"`

	// Error message
	Message string `json:"message"`
}

// Error makes it compatible with `error` interface.
func (he *ErrorMessage) Error() string {
	return fmt.Sprintf("code=%d, message=%v", he.Code, he.Message)
}
