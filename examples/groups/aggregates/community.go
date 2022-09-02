package aggregates

import (
	"context"
	"errors"
	"time"

	"github.com/contextcloud/eventstore/es"

	"github.com/contextcloud/eventstore/examples/groups/commands"
	"github.com/contextcloud/eventstore/examples/groups/events"
	"github.com/contextcloud/eventstore/examples/groups/models"
)

var (
	// ErrInvalidState protecting the aggregate
	ErrInvalidState = errors.New("Invalid state")

	// ErrForbidden when the user doesn't have perms to perform that action
	ErrForbidden = errors.New("Forbidden")

	// ErrAlreadyAdded you can't add a member who exists right?
	ErrAlreadyAdded = errors.New("When a member already exists")

	// ErrNotAdded you can't update a member who doesn't exist right?
	ErrNotAdded = errors.New("When a member doesn't exist")
)

const (
	RoleIdOwner     = "owner"
	RoleIdOrganizer = "organizer"
	RoleIdMod       = "mod"
)

func StaffIndex(staff []*models.StaffMemberModel, id string) int {
	for i, s := range staff {
		if s.Id == id {
			return i
		}
	}

	return -1
}

// Community maintains the state of a community
type Community struct {
	es.BaseAggregateSourced

	State     string
	Deleted   bool
	CreatedAt time.Time
	DeletedAt *time.Time

	Staff         []*models.StaffMemberModel
	PayoutCountry string

	// community settings
	GeneralSettings  *models.CommunityGeneralSettingsModel
	AdvancedSettings *models.CommunityAdvancedSettingsModel
}

// HandleCommand create events and validate based on such command
func (a *Community) HandleCommand(ctx context.Context, cmd es.Command) error {
	switch c := cmd.(type) {
	case *commands.CommunityNewCommand:
		return a.handleCommunityNewCommand(ctx, c)
	case *commands.CommunityDeleteCommand:
		return a.handleCommunityDeleteCommand(ctx, c)
	}
	return nil
}

func (a *Community) handleCommunityNewCommand(ctx context.Context, cmd *commands.CommunityNewCommand) error {
	if a.State != "" {
		return ErrInvalidState
	}

	a.Apply(ctx, &events.CommunityCreated{
		By:              cmd.By,
		Alias:           cmd.Alias,
		Name:            cmd.Name,
		Description:     cmd.Description,
		Regions:         cmd.Regions,
		Logo:            cmd.Logo,
		Banner:          cmd.Banner,
		Hidden:          cmd.Hidden,
		DonationEnabled: cmd.DonationEnabled,
		PaymentCountry:  cmd.PaymentCountry,
	})
	a.Apply(ctx, &events.CommunityStaffAdded{
		AccountId: cmd.By,
		RoleId:    RoleIdOwner,
	})
	return nil
}
func (a *Community) handleCommunityDeleteCommand(ctx context.Context, cmd *commands.CommunityDeleteCommand) error {
	a.Apply(ctx, &events.CommunityDeleted{})
	return nil
}

// ApplyEvent to auth
func (a *Community) ApplyEvent(ctx context.Context, event *es.Event) error {
	switch e := event.Data.(type) {
	case *events.CommunityCreated:
		return a.applyCommunityCreated(ctx, event, e)
	case *events.CommunityDeleted:
		return a.applyCommunityDeleted(ctx, event, e)

	case *events.CommunityStaffAdded:
		return a.applyCommunityStaffAdded(ctx, event, e)
	}
	return nil
}

func (a *Community) applyCommunityCreated(ctx context.Context, event *es.Event, data *events.CommunityCreated) error {
	a.GeneralSettings = &models.CommunityGeneralSettingsModel{
		Alias:                     data.Alias,
		Name:                      data.Name,
		Description:               data.Description,
		Regions:                   data.Regions,
		Logo:                      &data.Logo,
		Banner:                    &data.Banner,
		Hidden:                    data.Hidden,
		CommunityDonationsEnabled: data.DonationEnabled,
	}
	a.PayoutCountry = data.PaymentCountry
	if len(a.PayoutCountry) == 0 {
		a.PayoutCountry = "au"
	}
	a.CreatedAt = event.Timestamp
	return nil
}

func (a *Community) applyCommunityDeleted(ctx context.Context, event *es.Event, data *events.CommunityDeleted) error {
	a.Deleted = true
	a.DeletedAt = &event.Timestamp
	return nil
}

func (a *Community) applyCommunityStaffAdded(ctx context.Context, event *es.Event, data *events.CommunityStaffAdded) error {
	a.Staff = append(a.Staff, &models.StaffMemberModel{
		Id:        data.AccountId,
		RoleId:    data.RoleId,
		UpdatedAt: event.Timestamp,
	})
	return nil
}

// NewCommunity factory for creating the aggregate
func NewCommunity(id string) es.Entity {
	return &Community{
		// BaseAggregateSourced: es.NewBaseAggregateSourced(id, "Community"),
	}
}
