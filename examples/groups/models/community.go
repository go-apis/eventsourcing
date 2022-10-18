package models

import (
	"time"

	"github.com/google/uuid"
)

// StaffMemberModel a community staff member
type StaffMemberModel struct {
	Id        uuid.UUID
	RoleId    string
	UpdatedAt time.Time
}

// CommunityGeneralSettingsModel is a collection of general community settings
type CommunityGeneralSettingsModel struct {
	Name        string
	Description string
	// used for short url, must be unique
	Alias   string
	Regions []string
	Banner  *string
	Logo    *string
	// toggleable
	Hidden                      bool
	CommunityDonationsEnabled   bool
	CompetitionDonationsEnabled *bool
}

// CommunityAdvancedSettingsModel is a collection of general community settings
type CommunityAdvancedSettingsModel struct {
	PrizeExpiry int
	Forms       []string
	IRSTaxForms bool
}
