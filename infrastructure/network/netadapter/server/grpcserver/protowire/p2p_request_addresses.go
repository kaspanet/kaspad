package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_RequestAddresses) toAppMessage() (appmessage.Message, error) {
	protoGetAddresses := x.RequestAddresses
	subnetworkID, err := protoGetAddresses.SubnetworkID.toDomain()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgRequestAddresses{
		IncludeAllSubnetworks: protoGetAddresses.IncludeAllSubnetworks,
		SubnetworkID:          subnetworkID,
	}, nil
}

func (x *KaspadMessage_RequestAddresses) fromAppMessage(msgGetAddresses *appmessage.MsgRequestAddresses) error {
	x.RequestAddresses = &RequestAddressesMessage{
		IncludeAllSubnetworks: msgGetAddresses.IncludeAllSubnetworks,
		SubnetworkID:          domainSubnetworkIDToProto(msgGetAddresses.SubnetworkID),
	}
	return nil
}
