package perms

import (
	"context"
	"fmt"
	"testing"

	"github.com/contextcloud/eventstore/pkg/identity"
)

type user struct {
	Roles      []string
	Identities []identity.Identity
}

func (u *user) GetRoles() []string {
	return u.Roles
}
func (u *user) GetIdentities() []identity.Identity {
	return u.Identities
}
func (u *user) GetAudience() string {
	return "default"
}

func Test_Perms(t *testing.T) {
	m := NewManager()
	m.AddRule(
		"User:{set_name}@{set_id}",
		"owner,create,delete",
		"ApiToken:/User/{set_id}/*",
	)
	m.AddRule(
		"Role:{set_name}@{set_id}/owner",
		"owner",
		"ApiToken:/{set_name}/{set_id}/*",
	)

	data := []struct {
		actor        string
		relationship string
		object       string
		out          error
	}{
		{
			actor:        "User:Standard@123",
			relationship: "owner",
			object:       "ApiToken:/User/123/randomid",
			out:          nil,
		},
		{
			actor:        "Role:Department@123/owner",
			relationship: "owner",
			object:       "ApiToken:/Department/123/randomid",
			out:          nil,
		},
	}

	for i, d := range data {
		t.Run(fmt.Sprintf("perms-%d", i), func(t *testing.T) {
			err := m.Check(d.actor, d.relationship, d.object)
			if err != d.out {
				t.Error("Invalid result")
			}
		})
	}
}

func Test_Helper(t *testing.T) {
	m := NewManager()
	m.AddRule(
		"Identity:{set_name}@{set_id}",
		"owner",
		"ApiToken:/user/{set_id}/*",
	)
	m.AddRule(
		"Role:{set_name}@{set_id}/owner",
		"owner",
		"ApiToken:/{set_name}/{set_id}/*",
	)

	// create the user.
	ctx := context.Background()
	ctx = identity.SetUser(ctx, &user{
		Roles: []string{
			"SuperAdmin",
			"Department@345/owner",
		},
		Identities: []identity.Identity{
			{
				UserId:     "123",
				Username:   "user1",
				Connection: "Standard",
			},
			{
				UserId:     "456",
				Username:   "user2",
				Connection: "Standard",
			},
		},
	})

	data := []struct {
		relationship string
		object       string
		out          error
	}{
		{
			relationship: "owner",
			object:       "ApiToken:/user/123/*",
			out:          nil,
		},
		{
			relationship: "owner",
			object:       "ApiToken:/department/345/*",
			out:          nil,
		},
	}

	for i, d := range data {
		t.Run(fmt.Sprintf("rule-%d", i), func(t *testing.T) {
			err := CheckCurrent(ctx, m, d.relationship, d.object)
			if err != d.out {
				t.Error("Invalid result")
			}
		})
	}
}
