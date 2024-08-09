package commands

import (
	"github.com/go-apis/eventsourcing/es"
	"github.com/google/uuid"
)

type CommunityNewCommand struct {
	es.BaseCommand
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

type CommunityDeleteCommand struct {
	es.BaseCommand
}
