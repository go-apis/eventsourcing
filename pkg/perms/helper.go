package perms

import (
	"context"

	"github.com/contextcloud/eventstore/pkg/identity"
)

func CheckCurrent(ctx context.Context, manager Manager, relationship string, object string) error {
	user, ok := identity.FromContext(ctx)
	if !ok || user == nil {
		return manager.Check("Anonymous", relationship, object)
	}

	for _, role := range user.GetRoles() {
		// check if the role is ok!
		if err := manager.Check("Role:"+role, relationship, object); err == nil {
			return nil
		}
	}

	for _, identity := range user.GetIdentities() {
		if err := manager.Check("Identity:"+identity.Connection+"@"+identity.UserId, relationship, object); err == nil {
			return nil
		}
	}

	return ErrCheckFailed
}
