package models

import "time"

// StaffMemberModel a community staff member
type StaffMemberModel struct {
	Id        string
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
