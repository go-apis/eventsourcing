package commands

import "github.com/google/uuid"

type CommunityNewCommand struct {
	AggregateId     uuid.UUID `protobuf:"bytes,1,opt,name=aggregate_id,json=aggregateId,proto3" json:"aggregate_id,omitempty"`
	By              string    `protobuf:"bytes,2,opt,name=by,proto3" json:"by,omitempty"`
	Alias           string    `protobuf:"bytes,3,opt,name=alias,proto3" json:"alias,omitempty"`
	Name            string    `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
	Description     string    `protobuf:"bytes,5,opt,name=description,proto3" json:"description,omitempty"`
	Regions         []string  `protobuf:"bytes,6,rep,name=regions,proto3" json:"regions,omitempty"`
	Logo            string    `protobuf:"bytes,7,opt,name=logo,proto3" json:"logo,omitempty"`
	Banner          string    `protobuf:"bytes,8,opt,name=banner,proto3" json:"banner,omitempty"`
	Hidden          bool      `protobuf:"varint,9,opt,name=hidden,proto3" json:"hidden,omitempty"`
	DonationEnabled bool      `protobuf:"varint,10,opt,name=donation_enabled,json=donationEnabled,proto3" json:"donation_enabled,omitempty"`
	PaymentCountry  string    `protobuf:"bytes,11,opt,name=payment_country,json=paymentCountry,proto3" json:"payment_country,omitempty"`
}

func (x *CommunityNewCommand) GetAggregateId() uuid.UUID {
	return x.AggregateId
}

type CommunityDeleteCommand struct {
	AggregateId uuid.UUID `protobuf:"bytes,1,opt,name=aggregate_id,json=aggregateId,proto3" json:"aggregate_id,omitempty"`
}

func (x *CommunityDeleteCommand) GetAggregateId() uuid.UUID {
	return x.AggregateId
}
