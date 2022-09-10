package events

type CommunityCreated struct {
	By              string   `protobuf:"bytes,1,opt,name=by,proto3" json:"by,omitempty"`
	Alias           string   `protobuf:"bytes,2,opt,name=alias,proto3" json:"alias,omitempty"`
	Name            string   `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Description     string   `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	Regions         []string `protobuf:"bytes,5,rep,name=regions,proto3" json:"regions,omitempty"`
	Logo            string   `protobuf:"bytes,6,opt,name=logo,proto3" json:"logo,omitempty"`
	Banner          string   `protobuf:"bytes,7,opt,name=banner,proto3" json:"banner,omitempty"`
	Hidden          bool     `protobuf:"varint,8,opt,name=hidden,proto3" json:"hidden,omitempty"`
	DonationEnabled bool     `protobuf:"varint,9,opt,name=donation_enabled,json=donationEnabled,proto3" json:"donation_enabled,omitempty"`
	PaymentCountry  string   `protobuf:"bytes,10,opt,name=payment_country,json=paymentCountry,proto3" json:"payment_country,omitempty"`
}

type CommunityDeleted struct {
}

type CommunityStaffAdded struct {
	AccountId string `protobuf:"bytes,1,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty"`
	RoleId    string `protobuf:"bytes,2,opt,name=role_id,json=roleId,proto3" json:"role_id,omitempty"`
}
