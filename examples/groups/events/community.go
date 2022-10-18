package events

import "github.com/google/uuid"

type CommunityCreated struct {
	By              uuid.UUID `json:"by,omitempty"`
	Alias           string    `json:"alias,omitempty"`
	Name            string    `json:"name,omitempty"`
	Description     string    `json:"description,omitempty"`
	Regions         []string  `json:"regions,omitempty"`
	Logo            string    `json:"logo,omitempty"`
	Banner          string    `json:"banner,omitempty"`
	Hidden          bool      `json:"hidden,omitempty"`
	DonationEnabled bool      `json:"donation_enabled,omitempty"`
	PaymentCountry  string    `json:"payment_country,omitempty"`
}

type CommunityDeleted struct {
}

type CommunityStaffAdded struct {
	AccountId uuid.UUID `json:"account_id,omitempty"`
	RoleId    string    `json:"role_id,omitempty"`
}
