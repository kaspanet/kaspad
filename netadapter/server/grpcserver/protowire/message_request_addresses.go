package protowire

import (
	"github.com/kaspanet/kaspad/domainmessage"
)

func (x *KaspadMessage_RequestAddresses) toWireMessage() (domainmessage.Message, error) {
	protoGetAddresses := x.RequestAddresses
	subnetworkID, err := protoGetAddresses.SubnetworkID.toWire()
	if err != nil {
		return nil, err
	}

	return &domainmessage.MsgRequestAddresses{
		IncludeAllSubnetworks: protoGetAddresses.IncludeAllSubnetworks,
		SubnetworkID:          subnetworkID,
	}, nil
}

func (x *KaspadMessage_RequestAddresses) fromWireMessage(msgGetAddresses *domainmessage.MsgRequestAddresses) error {
	x.RequestAddresses = &RequestAddressesMessage{
		IncludeAllSubnetworks: msgGetAddresses.IncludeAllSubnetworks,
		SubnetworkID:          wireSubnetworkIDToProto(msgGetAddresses.SubnetworkID),
	}
	return nil
}
