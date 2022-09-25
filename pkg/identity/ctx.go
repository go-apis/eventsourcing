package identity

import (
	"context"
)

type key int

const identity key = 0

func SetUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, identity, user)
}

func FromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(identity).(User)
	if ok {
		return user, ok
	}
	return nil, false
}
