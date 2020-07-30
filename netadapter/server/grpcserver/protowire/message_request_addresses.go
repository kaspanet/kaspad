package protowire

import (
	"github.com/kaspanet/kaspad/wire"
)

func (x *KaspadMessage_RequestAddresses) toWireMessage() (wire.Message, error) {
	protoGetAddresses := x.RequestAddresses
	subnetworkID, err := protoGetAddresses.SubnetworkID.toWire()
	if err != nil {
		return nil, err
	}

	return &wire.MsgRequestAddresses{
		IncludeAllSubnetworks: protoGetAddresses.IncludeAllSubnetworks,
		SubnetworkID:          subnetworkID,
	}, nil
}

func (x *KaspadMessage_RequestAddresses) fromWireMessage(msgGetAddresses *wire.MsgRequestAddresses) error {
	x.RequestAddresses = &RequestAddressesMessage{
		IncludeAllSubnetworks: msgGetAddresses.IncludeAllSubnetworks,
		SubnetworkID:          wireSubnetworkIDToProto(msgGetAddresses.SubnetworkID),
	}
	return nil
}
